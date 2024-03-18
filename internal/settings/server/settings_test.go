package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const serverCfgPath = "../../test/server_config.yaml"

func TestNewSettings(t *testing.T) {
	t.Parallel()

	set, err := NewSettings(serverCfgPath)
	require.NoError(t, err)
	require.NotNil(t, set)

	expectedSet := Settings{}
	expectedSet.Database.URI = "database_uri"
	expectedSet.API.Address = "127.0.0.1"
	expectedSet.API.Port = 8080
	expectedSet.Log.Level = "debug"
	expectedSet.Log.Verbose = true
	expectedSet.Log.Format = "text"
	expectedSet.JWT.Secret = "jwt"
	expectedSet.JWT.Lifetime.Access = "1d"
	expectedSet.JWT.Lifetime.Refresh = "2d"
	expectedSet.Crypto.Key = "key"
	expectedSet.Crypto.Salt = "salt"

	expectedSet.Asymmetry.KeysPath = "./.cert"
	expectedSet.Asymmetry.PrivateKey = "private.pem"
	expectedSet.Asymmetry.PublicKey = "public.pem"

	require.Equal(t, expectedSet, *set)
}
