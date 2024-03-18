package token

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	ctx := context.Background()
	payload := Payload{
		ID: 100500,
	}

	ctx = ToContext(ctx, payload)

	newPayload, err := FromContext(ctx)
	require.NoError(t, err)

	assert.Equal(t, payload, newPayload)
}
