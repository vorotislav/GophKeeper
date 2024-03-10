package auth

import (
	"errors"
	"fmt"

	"GophKeeper/internal/settings/server"

	"github.com/xhit/go-str2duration/v2"
)

var (
	// errEmptySecret is returned when empty secret received in the settings.
	errEmptySecret = errors.New("authorizer empty JWT secret not allowed")
)

type Config struct {
	Secret               string
	AccessTokenLifetime  int
	RefreshTokenLifetime int
}

func NewConfig(s server.JwtSettings) (Config, error) {
	if s.Secret == "" {
		return Config{}, errEmptySecret
	}

	accessDuration, err := str2duration.ParseDuration(s.Lifetime.Access)
	if err != nil {
		return Config{}, fmt.Errorf("authorizer access token lifetime parsing failed: %w", err)
	}

	refreshDuration, err := str2duration.ParseDuration(s.Lifetime.Refresh)
	if err != nil {
		return Config{}, fmt.Errorf("authorizer refresh token lifetime parsing failed: %w", err)
	}

	return Config{
		Secret:               s.Secret,
		AccessTokenLifetime:  int(accessDuration),
		RefreshTokenLifetime: int(refreshDuration),
	}, nil
}
