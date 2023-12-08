// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package wasireceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wasireceiver"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental/gojs"
	"github.com/tetratelabs/wazero/experimental/sock"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

func NewPlugin(ctx context.Context, pluginPath string) (*Plugin, error) {
	p := Plugin{
		path: pluginPath,
	}
	runtime, module, err := p.instantiate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer module.Close(ctx)
	defer runtime.Close(ctx)

	err = p.initMetadata(ctx, module)
	if err != nil {
		return nil, fmt.Errorf("failed to init metadata: %w", err)
	}
	err = p.initDefaultConfig(ctx, module)
	if err != nil {
		return nil, fmt.Errorf("failed to init default config: %w", err)
	}
	err = p.initReceivers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init receivers: %w", err)
	}

	return &p, nil
}

func (p *Plugin) instantiate(ctx context.Context, moduleConfig wazero.ModuleConfig) (wazero.Runtime, api.Module, error) {
	wasm, err := os.ReadFile(p.path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read wasm plugin: %w", err)
	}

	runtime := wazero.NewRuntime(ctx)

	compiledModule, err := runtime.CompileModule(ctx, wasm)
	if err != nil {
		runtime.Close(ctx)
		return nil, nil, err
	}

	// Needed for WASI builds.
	if isWASIModule(compiledModule) {
		wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	}

	// Needed for Go WASM builds.
	if isGoWASMModule(compiledModule) {
		gojs.MustInstantiate(ctx, runtime, compiledModule)
	}

	if moduleConfig == nil {
		moduleConfig = wazero.NewModuleConfig().WithSysNanotime()
	}
	moduleCtx := sock.WithConfig(ctx, sock.NewConfig())
	module, err := runtime.InstantiateModule(moduleCtx, compiledModule, moduleConfig)
	if err != nil {
		runtime.Close(ctx)
		return nil, nil, fmt.Errorf("failed to instantiate wasm plugin: %w", err)
	}

	functions := module.ExportedFunctionDefinitions()
	expectedFunctions := []string{
		"metadata",
		"defaultConfig",

		"start",
		"stop",
	}
	for _, expectedFunction := range expectedFunctions {
		_, found := functions[expectedFunction]
		if !found {
			module.Close(ctx)
			runtime.Close(ctx)
			return nil, nil, fmt.Errorf("missing exported function %s", expectedFunction)
		}
	}

	return runtime, module, nil
}

type PluginMetadata struct {
	Type   component.Type       `mapstructure:"type"`
	Status PluginMetadataStatus `mapstructure:"status"`
}

type PluginMetadataStatus struct {
	Class         string              `mapstructure:"class"`
	Stability     map[string][]string `mastructure:"stability"`
	Distributions []string            `mapstructure:"distributions"`
	Codeowners    struct {
		Active []string `mapstructure:"active"`
	} `mapstructure:"codeowners"`
}

type Plugin struct {
	path string

	metadata      PluginMetadata
	defaultConfig any
	receivers     []receiver.FactoryOption
}

func stabilityLevel(name string) component.StabilityLevel {
	switch strings.ToLower(name) {
	case "undefined":
		return component.StabilityLevelUndefined
	case "unmaintained":
		return component.StabilityLevelUnmaintained
	case "deprecated":
		return component.StabilityLevelDeprecated
	case "development":
		return component.StabilityLevelDevelopment
	case "alpha":
		return component.StabilityLevelAlpha
	case "beta":
		return component.StabilityLevelBeta
	case "stable":
		return component.StabilityLevelStable
	}
	return 0
}

func (p *Plugin) Receivers() []receiver.FactoryOption {
	return p.receivers
}

func (p *Plugin) logReceiver(ctx context.Context) receiver.CreateLogsFunc {
	return func(ctx context.Context, settings receiver.CreateSettings, config component.Config, logs consumer.Logs) (receiver.Logs, error) {
		return &wasiLogsWrapper{
			plugin:   p,
			settings: settings,
			config:   config,
			logs:     logs,
		}, nil
	}
}

func (p *Plugin) DefaultConfig() component.Config {
	return component.Config(p.defaultConfig)
}

func (p *Plugin) NewReceiverFactory() receiver.Factory {
	return receiver.NewFactory(
		component.Type(p.metadata.Type),
		p.DefaultConfig,
		p.Receivers()...)
}

func (p *Plugin) initMetadata(ctx context.Context, module api.Module) error {
	var metadata PluginMetadata
	err := callJSONResponse(ctx, module, "metadata", &metadata)
	if err != nil {
		return err
	}
	p.metadata = metadata
	return nil
}

func (p *Plugin) initDefaultConfig(ctx context.Context, module api.Module) error {
	var defaultConfig map[string]any
	err := callJSONResponse(ctx, module, "defaultConfig", &defaultConfig)
	if err != nil {
		return err
	}
	p.defaultConfig = defaultConfig
	return nil
}

func (p *Plugin) initReceivers(ctx context.Context) error {
	var options []receiver.FactoryOption
	for stability, receivers := range p.metadata.Status.Stability {
		for _, receiverType := range receivers {
			var option receiver.FactoryOption
			switch receiverType {
			case "logs":
				option = receiver.WithLogs(p.logReceiver(ctx), stabilityLevel(stability))
			}
			if option != nil {
				options = append(options, option)
			}
		}
	}
	p.receivers = options
	return nil
}

func callJSONResponse(ctx context.Context, module api.Module, functionName string, response any) error {
	result, err := module.ExportedFunction(functionName).Call(ctx)
	if err != nil {
		return fmt.Errorf("failed to call %s: %w", functionName, err)
	}
	if len(result) == 0 || result[0] == 0 {
		// Nothing in the response.
		return nil
	}

	ptr := uint32(result[0] >> 32)
	size := uint32(result[0])

	d, _ := module.Memory().Read(uint32(ptr), uint32(size))
	if len(d) < int(size) {
		return fmt.Errorf("failed to read metadata from memory: out of range")
	}
	defer free(ctx, module, ptr)

	err = json.Unmarshal(d, &response)
	if err != nil {
		return fmt.Errorf("failed to decode metadata: %w", err)
	}

	return nil
}

func free(ctx context.Context, module api.Module, ptr uint32) {
	_, err := module.ExportedFunction("free").Call(ctx, uint64(ptr))
	if err != nil {
		panic(err)
	}
}

type wasiLogsWrapper struct {
	plugin   *Plugin
	settings receiver.CreateSettings
	config   component.Config
	logs     consumer.Logs

	runtime wazero.Runtime
	module  api.Module
}

type callResult struct {
	Error string `mapstructure:"error"`
}

func (w *wasiLogsWrapper) Start(ctx context.Context, host component.Host) error {
	runtime, module, err := w.plugin.instantiate(ctx, nil)
	if err != nil {
		return err
	}
	w.runtime = runtime
	w.module = module

	var result callResult
	err = callJSONResponse(ctx, w.module, "start", &result)
	if err != nil {
		return fmt.Errorf("failed to execute start: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("failed to start: %s", result.Error)
	}
	return nil
}

func (w *wasiLogsWrapper) Shutdown(ctx context.Context) error {
	defer w.runtime.Close(ctx)
	defer w.module.Close(ctx)

	var result callResult
	err := callJSONResponse(ctx, w.module, "stop", &result)
	if err != nil {
		return fmt.Errorf("failed to execute stop: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("failed to start: %s", result.Error)
	}

	return nil
}

func isWASIModule(module wazero.CompiledModule) bool {
	for _, f := range module.ImportedFunctions() {
		moduleName, _, _ := f.Import()
		switch moduleName {
		case "wasi_snapshot_preview1":
			return true
		}
	}
	return false
}

func isGoWASMModule(module wazero.CompiledModule) bool {
	for _, f := range module.ImportedFunctions() {
		moduleName, _, _ := f.Import()
		switch moduleName {
		case "go", "gojs":
			return true
		}
	}
	return false
}
