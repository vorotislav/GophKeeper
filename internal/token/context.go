// Package token упрощает работу с токеном в контексте.
package token

import (
	"context"
	"errors"
)

var (
	// ErrNotFound возвращается если токен не найден в контексте.
	ErrNotFound = errors.New("token payload not found in context")
	// ErrInvalidPayload возвращается если по искомому ключу в контексте хранится не токен.
	ErrInvalidPayload = errors.New("token payload is invalid")
)

// ctxTokenPayload структура для ключа хранения токена в контексте.
type ctxTokenPayload struct{}

// ToContext упаковывает токен в контекст.
func ToContext(ctx context.Context, payload Payload) context.Context {
	return context.WithValue(ctx, ctxTokenPayload{}, payload)
}

// FromContext достаёт токен из контекста.
func FromContext(ctx context.Context) (Payload, error) {
	value := ctx.Value(ctxTokenPayload{})
	if value == nil {
		return Payload{}, ErrNotFound
	}

	tokenPayload, ok := value.(Payload)
	if !ok {
		return Payload{}, ErrInvalidPayload
	}

	return tokenPayload, nil
}
