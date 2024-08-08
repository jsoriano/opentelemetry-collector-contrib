// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templatereceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent"
)

var (
	typeStr = component.MustNewType("template")

	receivers = sharedcomponent.NewSharedComponents()
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelAlpha),
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha),
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelAlpha),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createLogsReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Logs) (receiver.Logs, error) {
	return newTemplateReceiver(params, cfg.(*Config), consumer), nil
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return newTemplateReceiver(params, cfg.(*Config), consumer), nil
}

func createTracesReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Traces) (receiver.Traces, error) {
	return newTemplateReceiver(params, cfg.(*Config), consumer), nil
}
