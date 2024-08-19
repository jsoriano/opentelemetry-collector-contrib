// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templateprocessor

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	Name       string         `mapstructure:"name"`
	Pipeline   *component.ID  `mapstructure:"pipeline"`
	Version    string         `mapstructure:"version"`
	Parameters map[string]any `mapstructure:"parameters"`
}

func (cfg *Config) Validate() error {
	if cfg.Name == "" {
		return errors.New("name is required")
	}

	return nil
}
