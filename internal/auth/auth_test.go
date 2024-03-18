package auth

import (
	"testing"
	"time"

	"GophKeeper/internal/settings/server"
	"GophKeeper/internal/token"

	"github.com/stretchr/testify/assert"
)

func Test_Authorizer(t *testing.T) {
	sets := server.JwtSettings{ // valid settings
		Secret: "example",
		Lifetime: struct {
			Access  string `koanf:"access"`
			Refresh string `koanf:"refresh"`
		}{Access: "1m", Refresh: "2m"},
	}

	p := token.Payload{ID: 1}

	t.Run("valid", func(t *testing.T) {
		a, err := NewAuthorizer(sets)
		assert.NoError(t, err)

		s, err := a.GenerateToken(p)
		assert.NoError(t, err)
		assert.NotEqual(t, "", s)
		//fmt.Println(s) // print token for debug purposes

		ttl := a.GetRefreshTokenDurationLifetime()
		assert.Equal(t, 2*time.Minute, ttl)

		got, err := a.ParseToken(s)
		assert.NoError(t, err)
		assert.Equal(t, p, got)
	})

	t.Run("authorizer: invalid settings", func(t *testing.T) {
		_, err := NewAuthorizer(server.JwtSettings{})
		assert.Error(t, err)
	})

	t.Run("parse: expired token", func(t *testing.T) {
		ls := sets
		ls.Lifetime.Access = "0" // set zero TTL

		a, err := NewAuthorizer(ls)
		assert.NoError(t, err)

		s, err := a.GenerateToken(p)
		assert.NoError(t, err)
		assert.NotEqual(t, "", s)

		_, err = a.ParseToken(s)
		assert.ErrorIs(t, err, ErrTokenIsExpired)
	})

	t.Run("parse: invalid token", func(t *testing.T) {
		a, err := NewAuthorizer(sets)
		assert.NoError(t, err)

		_, err = a.ParseToken("invalid")
		assert.Error(t, err)
	})

	t.Run("parse: valid, but expired token", func(t *testing.T) {
		a, err := NewAuthorizer(sets)
		assert.NoError(t, err)

		_, err = a.ParseToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDYzNTU2NDgsImlhdCI6MTcwNjM1NTU4OCwiYWN0b3JfdHlwZSI6MSwiaWQiOjF9.Dn6lXpm8HiuCaD7Dqy4zaOOQTk2xZ7m_G5jPB5nJqxI")
		assert.ErrorIs(t, err, ErrTokenIsExpired)
	})

	t.Run("parse: valid, but unexpected signing method", func(t *testing.T) {
		a, err := NewAuthorizer(sets)
		assert.NoError(t, err)

		_, err = a.ParseToken("eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJmb28iOiJiYXIifQ.FhkiHkoESI_cG3NPigFrxEk9Z60_oXrOT2vGm9Pn6RDgYNovYORQmmA0zs1AoAOf09ly2Nx2YAg6ABqAYga1AcMFkJljwxTT5fYphTuqpWdy4BELeSYJx5Ty2gmr8e7RonuUztrdD5WfPqLKMm1Ozp_T6zALpRmwTIW0QPnaBXaQD90FplAg46Iy1UlDKr-Eupy0i5SLch5Q-p2ZpaL_5fnTIUDlxC3pWhJTyx_71qDI-mAA_5lE_VdroOeflG56sSmDxopPEG3bFlSu1eowyBfxtu0_CuVd-M42RU75Zc4Gsj6uV77MBtbMrf4_7M_NUTSgoIF3fRqxrj0NzihIBg")
		assert.ErrorIs(t, err, errSignMethod)
	})
}
