// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filetemplateextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/filetemplateextension"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

var (
	extensionType = component.MustNewType("file_templates")
)

const (
	extensionStability = component.StabilityLevelAlpha
)

// NewFactory creates a factory for ack extension.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		extensionType,
		createDefaultConfig,
		createExtension,
		extensionStability,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Path: "templates",
	}
}

func createExtension(_ context.Context, _ extension.Settings, cfg component.Config) (extension.Extension, error) {
	return newFileTemplateExtension(cfg.(*Config)), nil
}
