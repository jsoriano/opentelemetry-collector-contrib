// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"
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
