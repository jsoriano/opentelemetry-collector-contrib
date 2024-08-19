// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package templateprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

var (
	typeStr = component.MustNewType("template")
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithLogs(createLogsProcessor, component.StabilityLevelAlpha),
		processor.WithMetrics(createMetricsProcessor, component.StabilityLevelAlpha),
		processor.WithTraces(createTracesProcessor, component.StabilityLevelAlpha),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createLogsProcessor(_ context.Context, params processor.Settings, cfg component.Config, consumer consumer.Logs) (processor.Logs, error) {
	return newTemplateLogsProcessor(params, cfg.(*Config), consumer), nil
}

func createMetricsProcessor(_ context.Context, params processor.Settings, cfg component.Config, consumer consumer.Metrics) (processor.Metrics, error) {
	return newTemplateMetricsProcessor(params, cfg.(*Config), consumer), nil
}

func createTracesProcessor(_ context.Context, params processor.Settings, cfg component.Config, consumer consumer.Traces) (processor.Traces, error) {
	return newTemplateTracesProcessor(params, cfg.(*Config), consumer), nil
}
