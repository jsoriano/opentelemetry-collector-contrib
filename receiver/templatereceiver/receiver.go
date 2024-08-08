// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

type templateReceiver[T any] struct {
	params       receiver.Settings
	config       *Config
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
	return nil
}

func (r *templateReceiver[T]) Shutdown(ctx context.Context) error {
	return nil
}
