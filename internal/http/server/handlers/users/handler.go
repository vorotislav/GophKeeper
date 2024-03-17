// Package users реализует HTTP обработчик для /login и /register
package users

import (
	"context"
	"encoding/json"
	"errors"

	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"

	"go.uber.org/zap"
	"net/http"
)

// UserProvider описывает методы для работы с пользователями.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=UserProvider --with-expecter=true
type UserProvider interface {
	UserCreate(ctx context.Context, um models.UserMachine) (models.Session, error)
	UserLogin(ctx context.Context, um models.UserMachine) (models.Session, error)
	//UserLogout(ctx context.Context, um models.UserMachine) error
}

// Handler реализует HTTP обработчик для /login и /register.
type Handler struct {
	log          *zap.Logger
	userProvider UserProvider
}

// NewHandler конструктор для Handler.
func NewHandler(log *zap.Logger, up UserProvider) *Handler {
	return &Handler{
		log:          log.Named("user handler"),
		userProvider: up,
	}
}

// Register обрабатывает POST запрос на создание пользователя /register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var um models.UserMachine

	if err := json.NewDecoder(r.Body).Decode(&um); err != nil {
		h.log.Error("failed to register user decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed to register user decode", err.Error()))

		return
	}

	res, err := h.userProvider.UserCreate(r.Context(), um)

	if err != nil {
		h.log.Error("failed to user register", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed to register user", err.Error()))

		return
	}

	responder.JSON(w, &output{
		session:    res,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success user register")
}

// Login обрабатывает POST запрос на входа пользователя /login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var um models.UserMachine

	if err := json.NewDecoder(r.Body).Decode(&um); err != nil {
		h.log.Error("failed to login user decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed to login user decode", err.Error()))

		return
	}

	res, err := h.userProvider.UserLogin(r.Context(), um)

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Error("failed to user login", zap.Error(err))

			responder.JSON(w, httpErr.NewNotFoundError("user not found"))

			return
		}

		h.log.Error("failed to user login", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed to login user", err.Error()))

		return
	}

	responder.JSON(w, &output{
		session:    res,
		statusCode: http.StatusOK,
	})

	h.log.Info("success user login")
}

type input struct {
	User struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"user"`
	Machine struct {
		IPAddress string `json:"ip_address"`
	} `json:"machine"`
}

type output struct {
	session    models.Session
	statusCode int
}

// ToJSON converts output structure into JSON representation.
func (o *output) ToJSON() ([]byte, error) { return json.Marshal(o.session) }

// StatusCode allows to customize output HTTP status code (when responder.JSON is used).
func (o *output) StatusCode() int { return o.statusCode }
