package client

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const serverCfgPath = "../../test/client_config.yaml"

func TestNewSettings(t *testing.T) {
	t.Parallel()

	set, err := NewSettings(serverCfgPath)
	require.NoError(t, err)
	require.NotNil(t, set)

	expectedSet := Settings{}
	expectedSet.Server.Address = "127.0.0.1:8080"
	expectedSet.Log.Level = "debug"
	expectedSet.Log.Verbose = true
	expectedSet.Log.Format = "json"

	expectedSet.Asymmetry.KeysPath = "./.cert"
	expectedSet.Asymmetry.PrivateKey = "private.pem"
	expectedSet.Asymmetry.PublicKey = "public.pem"

	require.Equal(t, expectedSet, *set)
}
