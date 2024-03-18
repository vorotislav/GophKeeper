package models

import (
	"errors"
)

// описаны глобальные ошибки, которые могут проходить сквозь слои приложения.
var (
	// ErrNotFound возвращается, когда объект\ресурс или пользователь не найден.
	ErrNotFound = errors.New("not found")
	// ErrInvalidPassword возвращается, когда была предпринята попытка войти с неправильным паролем.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidInput возвращается, когда было обнаружено, что данные некорректные.
	ErrInvalidInput = errors.New("invalid input")
)

const (
	// InputTimeFormLong описывает формат вывода времени, используемый во всем приложении.
	InputTimeFormLong = "2006-01-02 15:04:05"
)
