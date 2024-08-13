// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"
	"fmt"
	"path/filepath"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
)

type template interface {
	uri() string
	providerFactory() confmap.ProviderFactory
}

func findTemplate(ctx context.Context, host component.Host, name string, version string) (template, error) {
	// TODO: use extensions to provide collections of templates instead of hard-coding path.
	// TODO: versioning
	templatePath := filepath.Join("templates", name+".yml")
	return &templateFile{
		path: templatePath,
	}, nil
}

type templateFile struct {
	path string
}

func (t *templateFile) uri() string {
	return "file:" + t.path
}

func (t *templateFile) providerFactory() confmap.ProviderFactory {
	return fileprovider.NewFactory()
}

type templateConfig struct {
	Receivers  map[component.ID]map[string]any   `mapstructure:"receivers"`
	Processors map[component.ID]map[string]any   `mapstructure:"processors"`
	Pipelines  map[component.ID]templatePipeline `mapstructure:"pipelines"`
}

func (c *templateConfig) Validate() error {
	for id, pipeline := range c.Pipelines {
		if err := pipeline.Validate(); err != nil {
			return fmt.Errorf("invalid pipeline %q: %w", id, err)
		}

		if _, found := c.Receivers[pipeline.Receiver]; !found {
			return fmt.Errorf("receiver %s not defined", pipeline.Receiver.String())
		}

		for _, processor := range pipeline.Processors {
			if _, found := c.Processors[processor]; !found {
				return fmt.Errorf("processor %q not defined", processor.String())
			}

		}
	}

	return nil
}

type templatePipeline struct {
	Receiver   component.ID   `mapstructure:"receiver"`
	Processors []component.ID `mapstructure:"processors"`
}

func (p *templatePipeline) Validate() error {
	if len(p.Receiver.Type().String()) == 0 {
		return fmt.Errorf("pipeline without receiver")
	}

	return nil
}
