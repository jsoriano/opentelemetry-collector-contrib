// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

type templateReceiver struct {
	params              receiver.Settings
	config              *Config
	nextLogsConsumer    consumer.Logs
	nextMetricsConsumer consumer.Metrics
	nextTracesConsumer  consumer.Traces
}

func newTemplateReceiver(params receiver.Settings, config *Config) component.Component {
	return &templateReceiver{
		params: params,
		config: config,
	}
}

func (r *templateReceiver) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (r *templateReceiver) Shutdown(ctx context.Context) error {
	return nil
}
