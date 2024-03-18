package repository

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestNewRepository(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	validURI := "postgres://postgres@127.0.0.1:5432/gophkeeper_test"
	repo, err := NewRepository(context.Background(), log, validURI)
	require.NoError(t, err)
	require.NotNil(t, repo)

	invalidURI := "invalid uri"
	repo2, err := NewRepository(context.Background(), log, invalidURI)
	require.Error(t, err)
	require.Nil(t, repo2)
}
