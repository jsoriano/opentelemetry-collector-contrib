// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import "errors"

type Config struct {
	Name       string         `mapstructure:"name"`
	Pipelines  []string       `mapstructure:"pipelines"`
	Version    string         `mapstructure:"version"`
	Parameters map[string]any `mapstructure:"parameters"`
}

func (cfg *Config) Validate() error {
	if cfg.Name == "" {
		return errors.New("name is required")
	}

	return nil
}
