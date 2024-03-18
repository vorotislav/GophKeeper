package auth

import (
	"errors"
	"fmt"
	"time"

	"GophKeeper/internal/settings/server"
	"GophKeeper/internal/token"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xhit/go-str2duration/v2"
)

var (
	// ErrTokenIsExpired is returned when token lifetime is expired.
	ErrTokenIsExpired = errors.New("token TTL is expired")

	// errParseToken is returned when token parse failed.
	errParseToken = errors.New("parse token string")

	// errSignMethod is returned when token parse failed due to unexpected signing method.
	errSignMethod = errors.New("unexpected signing method")

	// errEmptySecret is returned when empty secret received in the settings.
	errEmptySecret = errors.New("authorizer empty JWT secret not allowed")
)

// LeewayDuration defines the leeway for matching NotBefore/Expiry claims.
const LeewayDuration = 5

// Claims contains internal claims and user payload.
type Claims struct {
	jwt.RegisteredClaims
	token.Payload
}

type Authorizer struct {
	Secret               string
	AccessTokenLifetime  int
	RefreshTokenLifetime int
}

// NewAuthorizer returns Authorizer.
func NewAuthorizer(settings server.JwtSettings) (*Authorizer, error) {
	if settings.Secret == "" {
		return nil, errEmptySecret
	}

	accessDuration, err := str2duration.ParseDuration(settings.Lifetime.Access)
	if err != nil {
		return nil, fmt.Errorf("authorizer access token lifetime parsing failed: %w", err)
	}

	refreshDuration, err := str2duration.ParseDuration(settings.Lifetime.Refresh)
	if err != nil {
		return nil, fmt.Errorf("authorizer refresh token lifetime parsing failed: %w", err)
	}

	return &Authorizer{
		Secret:               settings.Secret,
		AccessTokenLifetime:  int(accessDuration),
		RefreshTokenLifetime: int(refreshDuration),
	}, nil
}

func (a *Authorizer) GetRefreshTokenDurationLifetime() time.Duration {
	return time.Duration(a.RefreshTokenLifetime)
}

func (a *Authorizer) GenerateToken(payload token.Payload) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(a.AccessTokenLifetime))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "",
		},
		Payload: payload,
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := tkn.SignedString([]byte(a.Secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return ss, nil
}

func (a *Authorizer) ParseToken(tokenString string) (token.Payload, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", errSignMethod, token.Header["alg"])
		}

		return []byte(a.Secret), nil
	}, jwt.WithLeeway(LeewayDuration*time.Second))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return token.Payload{}, ErrTokenIsExpired
		}

		return token.Payload{}, fmt.Errorf("%w: %w", errParseToken, err)
	}

	if claims, ok := parsedToken.Claims.(*Claims); ok && parsedToken.Valid {
		if time.Now().After(claims.ExpiresAt.Time) {
			return token.Payload{}, ErrTokenIsExpired
		}

		return claims.Payload, nil
	}

	return token.Payload{}, fmt.Errorf("%w: %w", errParseToken, err)
}
