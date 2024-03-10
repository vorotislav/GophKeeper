package models

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidInput         = errors.New("invalid input")
	ErrForbidden            = errors.New("forbidden")
	ErrConflict             = errors.New("conflict")
	ErrUnprocessableContent = errors.New("unprocessable content")
)
