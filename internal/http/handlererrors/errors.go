package handlererrors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HandlerErrorCode string

func (e HandlerErrorCode) String() string { return string(e) }

const (
	ErrAppCode      HandlerErrorCode = "ERR_APP_CODE"
	ErrInvalidInput HandlerErrorCode = "ERR_INVALID_INPUT"
	ErrNotFound     HandlerErrorCode = "NOT_FOUND"
)

type HandlerError struct {
	Code           string `json:"code"`
	Title          string `json:"title"`
	Detail         string `json:"detail"`
	httpStatusCode int
}

func NewNotFoundError(title string) HandlerError {
	return HandlerError{
		Code:           string(ErrNotFound),
		Title:          title,
		Detail:         "",
		httpStatusCode: http.StatusNotFound,
	}
}

func NewInvalidInput(title, detail string) HandlerError {
	return HandlerError{
		Code:           string(ErrInvalidInput),
		Title:          title,
		Detail:         detail,
		httpStatusCode: http.StatusBadRequest,
	}
}

func NewInternalError(title, detail string) HandlerError {
	return HandlerError{
		Code:           string(ErrAppCode),
		Title:          title,
		Detail:         detail,
		httpStatusCode: http.StatusInternalServerError,
	}
}

func (he HandlerError) Error() string {
	return fmt.Sprintf("%s: %s: %s", he.Code, he.Title, he.Detail)
}

func (he HandlerError) StatusCode() int {
	return he.httpStatusCode
}

func (he HandlerError) ToJSON() ([]byte, error) { return json.Marshal(he) }
