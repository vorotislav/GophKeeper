package token

import (
	"context"
	"errors"
)

var (
	ErrNotFound       = errors.New("token payload not found in context")
	ErrInvalidPayload = errors.New("token payload is invalid")
)

type ctxTokenPayload struct{}

func ToContext(ctx context.Context, payload Payload) context.Context {
	return context.WithValue(ctx, ctxTokenPayload{}, payload)
}

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
