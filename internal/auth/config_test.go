package auth

import (
	"testing"

	"GophKeeper/internal/settings/server"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	testcases := []struct {
		name  string
		valid bool
		sets  server.JwtSettings
	}{
		{
			name:  "valid",
			valid: true,
			sets: server.JwtSettings{
				Secret: "example",
				Lifetime: struct {
					Access  string `koanf:"access"`
					Refresh string `koanf:"refresh"`
				}{Access: "1m", Refresh: "1m"},
			},
		},
		{
			name:  "invalid: no secret",
			valid: false,
			sets:  server.JwtSettings{},
		},
		{
			name:  "invalid: no access ttl",
			valid: false,
			sets: server.JwtSettings{
				Secret: "example",
				Lifetime: struct {
					Access  string `koanf:"access"`
					Refresh string `koanf:"refresh"`
				}{Access: "", Refresh: "1m"},
			},
		},
		{
			name:  "invalid: no refresh ttl",
			valid: false,
			sets: server.JwtSettings{
				Secret: "example",
				Lifetime: struct {
					Access  string `koanf:"access"`
					Refresh string `koanf:"refresh"`
				}{Access: "1m", Refresh: ""},
			},
		},
	}

	for _, tc := range testcases {
		got, err := NewConfig(tc.sets)
		if tc.valid {
			assert.NoError(t, err)
			assert.NotNil(t, got)
		} else {
			assert.Error(t, err)
		}
	}
}
