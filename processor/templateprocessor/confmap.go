// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templateprocessor

import (
	"context"
	"strings"

	"go.opentelemetry.io/collector/confmap"
)

const (
	varProviderScheme = "var"
)

func newVariablesProviderFactory(variables map[string]any) confmap.ProviderFactory {
	return confmap.NewProviderFactory(createVariablesProvider(variables))
}

func createVariablesProvider(variables map[string]any) confmap.CreateProviderFunc {
	return func(_ confmap.ProviderSettings) confmap.Provider {
		return &variablesProvider{variables: variables}
	}
}

type variablesProvider struct {
	variables map[string]any
}

func (p *variablesProvider) Retrieve(ctx context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	key := strings.TrimLeft(uri, varProviderScheme+":")
	value, found := p.variables[key]
	if !found {
		// FIXME: Resolve relevant configuration only instead of the whole template, so we don't need to ignore
		// variables not found in unrelevant parts.
		//return nil, fmt.Errorf("variable %q not found", key)
		return confmap.NewRetrieved("")
	}

	return confmap.NewRetrieved(value)
}

func (p *variablesProvider) Scheme() string {
	return varProviderScheme
}

func (p *variablesProvider) Shutdown(ctx context.Context) error {
	return nil
}
