// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("not found")

type Template interface {
	URI() string
	ProviderFactory() confmap.ProviderFactory
}

type TemplateFinder interface {
	FindTemplate(ctx context.Context, name, version string) (Template, error)
}

func (r *templateReceiver) findTemplate(ctx context.Context, host component.Host, name string, version string) (Template, error) {
	anyExtension := false
	for eid, extension := range host.GetExtensions() {
		finder, ok := extension.(TemplateFinder)
		if !ok {
			continue
		}
		anyExtension = true

		template, err := finder.FindTemplate(ctx, name, version)
		if errors.Is(ErrNotFound, err) {
			continue
		}
		if err != nil {
			r.params.Logger.Error("template finder failed",
				zap.String("component", eid.String()),
				zap.Error(err))
			return nil, err
		}

		return template, nil
	}
	if !anyExtension {
		return nil, errors.New("no template finder extension found")
	}

	return nil, ErrNotFound
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
