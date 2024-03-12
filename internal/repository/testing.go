package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

const (
	testConninfo = "postgres://postgres@127.0.0.1:5432/gophkeeper_test"
)

// TestRepository create new repository for testing purposes.
func TestRepository(t *testing.T) *Repo {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	repo, err := NewRepository(context.Background(), log, testConninfo)
	assert.NoError(t, err)

	return repo
}
