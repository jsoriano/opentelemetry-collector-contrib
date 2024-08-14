// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filetemplateextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/filetemplateextension"

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/templatereceiver"
)

type fileTemplateExtension struct {
	config *Config
}

var _ templatereceiver.TemplateFinder = &fileTemplateExtension{}

func newFileTemplateExtension(config *Config) *fileTemplateExtension {
	return &fileTemplateExtension{
		config: config,
	}
}

func (e *fileTemplateExtension) FindTemplate(ctx context.Context, name, version string) (templatereceiver.Template, error) {
	path := filepath.Join(e.config.Path, name+".yml")
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, templatereceiver.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &templateFile{
		path: path,
	}, nil
}

func (*fileTemplateExtension) Start(context.Context, component.Host) error {
	return nil
}

func (*fileTemplateExtension) Shutdown(context.Context) error {
	return nil
}

type templateFile struct {
	path string
}

func (t *templateFile) URI() string {
	return "file:" + t.path
}

func (t *templateFile) ProviderFactory() confmap.ProviderFactory {
	return fileprovider.NewFactory()
}
