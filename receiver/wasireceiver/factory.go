// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package wasireceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wasireceiver"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

func NewPlugin(ctx context.Context, pluginPath string) (*Plugin, error) {
	wasm, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm plugin: %w", err)
	}

	runtime := wazero.NewRuntime(ctx)

	// Needed for TinyGo builds.
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	module, err := runtime.Instantiate(ctx, wasm)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate wasm plugin: %w", err)
	}

	functions := module.ExportedFunctionDefinitions()
	expectedFunctions := []string{
		"metadata",
		"defaultConfig",
	}
	for _, expectedFunction := range expectedFunctions {
		_, found := functions[expectedFunction]
		if !found {
			return nil, fmt.Errorf("missing exported function %s", expectedFunction)
		}
	}

	metadata, err := getMetadata(ctx, module)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return &Plugin{
		runtime:  runtime,
		module:   module,
		metadata: metadata,

		defaultConfig: struct{}{},
	}, nil
}

type PluginMetadata struct {
	Type string `mapstructure:"type"`
}

type Plugin struct {
	runtime wazero.Runtime
	module  api.Module

	metadata      PluginMetadata
	defaultConfig any
}

func (p *Plugin) Close(ctx context.Context) error {
	return p.runtime.Close(ctx)
}

func (p *Plugin) Receivers() []receiver.FactoryOption {
	return nil
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

func getMetadata(ctx context.Context, module api.Module) (PluginMetadata, error) {
	result, err := module.ExportedFunction("metadata").Call(ctx)
	if err != nil {
		return PluginMetadata{}, fmt.Errorf("failed to retrieve metadata: %w", err)
	}
	ptr := uint32(result[0] >> 32)
	size := uint32(result[0])

	d, _ := module.Memory().Read(uint32(ptr), uint32(size))
	if len(d) < int(size) {
		return PluginMetadata{}, fmt.Errorf("failed to read metadata from memory: out of range")
	}
	defer free(ctx, module, ptr)

	var metadata PluginMetadata
	err = json.Unmarshal(d, &metadata)
	if err != nil {
		return metadata, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return metadata, nil
}

func free(ctx context.Context, module api.Module, ptr uint32) {
	_, err := module.ExportedFunction("free").Call(ctx, uint64(ptr))
	if err != nil {
		panic(err)
	}
}
