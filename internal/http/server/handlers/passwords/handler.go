// Package passwords реализует HTTP обработчик для /passwords
package passwords

import (
	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// PasswordProvider описывает методы для работы с паролями пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=PasswordProvider --with-expecter=true
type PasswordProvider interface {
	PasswordCreate(ctx context.Context, p models.Password) error
	PasswordUpdate(ctx context.Context, p models.Password) error
	PasswordDelete(ctx context.Context, id int) error
	Passwords(ctx context.Context) ([]models.Password, error)
}

// Handler реализует HTTP обработчик для /passwords.
type Handler struct {
	log              *zap.Logger
	passwordProvider PasswordProvider
}

// NewHandler конструктор для Handler.
func NewHandler(log *zap.Logger, pp PasswordProvider) *Handler {
	return &Handler{
		log:              log.Named("passwords handler"),
		passwordProvider: pp,
	}
}

// PasswordCreate обрабатывает POST запрос на создание пароля.
func (h *Handler) PasswordCreate(w http.ResponseWriter, r *http.Request) {
	var (
		pass models.Password
		err  error
	)

	if err := json.NewDecoder(r.Body).Decode(&pass); err != nil {
		h.log.Error("failed password decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed password decode", err.Error()))

		return
	}

	err = h.passwordProvider.PasswordCreate(r.Context(), pass)

	if err != nil {
		h.log.Error("failed password create", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed password create", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed password create", err.Error()))

		return
	}

	setHeaders(w, http.StatusCreated)

	h.log.Debug("success password create")
}

// PasswordUpdate обрабатывает PUT запрос на обновление пароля.
func (h *Handler) PasswordUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		pass models.Password
		err  error
	)

	if err := json.NewDecoder(r.Body).Decode(&pass); err != nil {
		h.log.Error("failed password decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed password decode", err.Error()))

		return
	}

	err = h.passwordProvider.PasswordUpdate(r.Context(), pass)

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("passwords not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed password update"))

			return
		}

		h.log.Error("failed password update", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed password update", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed password update", err.Error()))

		return
	}

	setHeaders(w, http.StatusAccepted)
}

// PasswordDelete обрабатывает DELETE запрос на удаление пароля.
func (h *Handler) PasswordDelete(w http.ResponseWriter, r *http.Request) {
	strPasswordID := chi.URLParam(r, "passwordID")
	passwordID, err := strconv.Atoi(strPasswordID)
	if err != nil {
		h.log.Error("failed get password id", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed password delete", err.Error()))

		return
	}

	err = h.passwordProvider.PasswordDelete(r.Context(), passwordID)
	if err != nil {
		h.log.Error("failed password delete", zap.Error(err))

		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("passwords not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed password delete"))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed password delete", err.Error()))

		return
	}

	setHeaders(w, http.StatusNoContent)
}

// Passwords обрабатывает GET запрос на получение паролей.
func (h *Handler) Passwords(w http.ResponseWriter, r *http.Request) {
	passs, err := h.passwordProvider.Passwords(r.Context())
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("passwords not found")

			http.Error(w, fmt.Sprintf("failed get passwords: %s", err.Error()), http.StatusNotFound)

			return
		}

		h.log.Error("failed get passwords", zap.Error(err))

		http.Error(w, fmt.Sprintf("failed get passwords: %s", err.Error()), http.StatusInternalServerError)

		return
	}

	if len(passs) == 0 {
		http.Error(w, fmt.Sprintf("failed get passwords"), http.StatusNotFound)

		return
	}

	responder.JSON(w, &output{
		items:      passs,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get passwords")
}

type output struct {
	items      []models.Password
	statusCode int
}

// ToJSON converts output structure into JSON representation.
func (o *output) ToJSON() ([]byte, error) { return json.Marshal(o.items) }

// StatusCode allows to customize output HTTP status code (when responder.JSON is used).
func (o *output) StatusCode() int { return o.statusCode }

func setHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}
