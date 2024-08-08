// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/receiver"
)

type templateReceiver[T any] struct {
	params       receiver.Settings
	config       *Config
	component    component.Component
	nextConsumer T
}

func newTemplateReceiver[T any](params receiver.Settings, config *Config, consumer T) *templateReceiver[T] {
	return &templateReceiver[T]{
		params:       params,
		config:       config,
		nextConsumer: consumer,
	}
}

func (r *templateReceiver[T]) Start(ctx context.Context, host component.Host) error {
	template, err := findTemplate(ctx, host, r.config.Name, r.config.Version)
	if err != nil {
		return fmt.Errorf("failed to find template %q: %w", r.config.Name, err)
	}

	resolver, err := newResolver(template, r.config.Parameters)
	if err != nil {
		return fmt.Errorf("failed to create configuration resolver: %w", err)
	}
	_, err = resolver.Resolve(ctx)
	if err != nil {
		return fmt.Errorf("failed to resolve template: %w", err)
	}

	return nil
}

func (r *templateReceiver[T]) Shutdown(ctx context.Context) error {
	if r.component == nil {
		return nil
	}

	return r.component.Shutdown(ctx)
}

func newResolver(template template, variables map[string]any) (*confmap.Resolver, error) {
	settings := confmap.ResolverSettings{
		URIs: []string{template.uri()},
		ProviderFactories: []confmap.ProviderFactory{
			template.providerFactory(),
			newVariablesProviderFactory(variables),
		},
	}
	return confmap.NewResolver(settings)
}
