package wasireceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

func TestNewPlugin(t *testing.T) {
	p, err := NewPlugin(context.Background(), "./testdata/test.wasm")
	require.NoError(t, err)
	defer func() {
		err = p.Close(context.Background())
		assert.NoError(t, err)
	}()

	assert.Equal(t, "test", string(p.metadata.Type))
	assert.EqualValues(t, map[string][]string{"development": {"logs"}}, p.metadata.Status.Stability)
	assert.Equal(t, map[string]any{}, p.defaultConfig)
}

func TestFactory(t *testing.T) {
	p, err := NewPlugin(context.Background(), "./testdata/test.wasm")
	require.NoError(t, err)
	defer p.Close(context.Background())

	factories, err := receiver.MakeFactoryMap(
		p.NewReceiverFactory(),
	)
	assert.NoError(t, err)

	config := map[component.ID]component.Config{
		component.NewID(p.metadata.Type): p.defaultConfig,
	}
	builder := receiver.NewBuilder(config, factories)
	assert.NotNil(t, builder)
}
