// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filetemplateextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/filetemplateextension"

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Path string `mapstructure:"path"`
}

func (c *Config) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}
	info, err := os.Stat(c.Path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path (%s) must be a directory", c.Path)
	}

	return nil
}
