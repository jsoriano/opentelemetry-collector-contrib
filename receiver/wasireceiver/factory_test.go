package wasireceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

func TestNewPlugin(t *testing.T) {
	p, err := NewPlugin(context.Background(), "./testdata/test.wasm")
	require.NoError(t, err)

	assert.Equal(t, "test", string(p.metadata.Type))
	assert.EqualValues(t, map[string][]string{"development": {"logs"}}, p.metadata.Status.Stability)
	assert.Equal(t, map[string]any{}, p.defaultConfig)
}

func TestFactory(t *testing.T) {
	ctx := context.Background()

	p, err := NewPlugin(ctx, "./testdata/test.wasm")
	require.NoError(t, err)

	factories, err := receiver.MakeFactoryMap(
		p.NewReceiverFactory(),
	)
	assert.NoError(t, err)

	config := map[component.ID]component.Config{
		component.NewID(p.metadata.Type): p.defaultConfig,
	}
	builder := receiver.NewBuilder(config, factories)
	assert.NotNil(t, builder)

	settings := receiver.CreateSettings{
		ID: component.NewID(p.metadata.Type),
	}
	settings.Logger = zap.Must(zap.NewDevelopment())

	_, err = builder.CreateLogs(ctx, settings, &testLogsConsumer{})
	assert.NoError(t, err)
}

type testLogsConsumer struct{}

func (*testLogsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (*testLogsConsumer) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	return nil
}
