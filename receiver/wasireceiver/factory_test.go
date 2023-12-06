package wasireceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlugin(t *testing.T) {
	p, err := NewPlugin(context.Background(), "./testdata/test.wasm")
	require.NoError(t, err)

	assert.Equal(t, "test", p.metadata.Type)
	assert.Equal(t, map[string]any{}, p.defaultConfig)
}
