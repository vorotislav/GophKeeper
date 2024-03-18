// Package cards реализует HTTP обработчик для /cards
package cards

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CardProvider описывает методы для работы с картами пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=CardProvider --with-expecter=true
type CardProvider interface {
	CardCreate(ctx context.Context, c models.Card) error
	CardUpdate(ctx context.Context, c models.Card) error
	CardDelete(ctx context.Context, id int) error
	Cards(ctx context.Context) ([]models.Card, error)
}

// Handler реализует HTTP обработчик для /cards.
type Handler struct {
	log          *zap.Logger
	cardProvider CardProvider
}

// NewHandler конструктор для Handler.
func NewHandler(log *zap.Logger, cp CardProvider) *Handler {
	return &Handler{
		log:          log.Named("cards handler"),
		cardProvider: cp,
	}
}

// CardCreate обрабатывает POST запрос на создание карты.
func (h *Handler) CardCreate(w http.ResponseWriter, r *http.Request) {
	var (
		card models.Card
		err  error
	)

	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		h.log.Error("failed card decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed card decode", err.Error()))

		return
	}

	err = h.cardProvider.CardCreate(r.Context(), card)

	if err != nil {
		h.log.Error("failed card create", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed card create", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed card create", err.Error()))

		return
	}

	setHeaders(w, http.StatusCreated)

	h.log.Info("success card create")
}

// CardUpdate обрабатывает PUT запрос на обновление карты.
func (h *Handler) CardUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		card models.Card
		err  error
	)

	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		h.log.Error("failed card decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed card decode", err.Error()))

		return
	}

	err = h.cardProvider.CardUpdate(r.Context(), card)

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("card not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed card update"))

			return
		}

		h.log.Error("failed card update", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed card update", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed card update", err.Error()))

		return
	}

	setHeaders(w, http.StatusAccepted)
}

// CardDelete обрабатывает DELETE запрос на удаление карты.
func (h *Handler) CardDelete(w http.ResponseWriter, r *http.Request) {
	strCardID := chi.URLParam(r, "cardID")
	cardID, err := strconv.Atoi(strCardID)
	if err != nil {
		h.log.Error("failed get card id", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed card delete", err.Error()))

		return
	}

	err = h.cardProvider.CardDelete(r.Context(), cardID)
	if err != nil {
		h.log.Error("failed card delete", zap.Error(err))

		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("card not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed card delete"))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed card delete", err.Error()))

		return
	}

	setHeaders(w, http.StatusNoContent)
}

// Cards обрабатывает GET запрос на получение карт.
func (h *Handler) Cards(w http.ResponseWriter, r *http.Request) {
	cards, err := h.cardProvider.Cards(r.Context())
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("cards not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed get cards"))

			return
		}

		h.log.Error("failed get cards", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed get cards", err.Error()))

		return
	}

	if len(cards) == 0 {
		responder.JSON(w, httpErr.NewNotFoundError("failed get cards"))

		return
	}

	responder.JSON(w, &output{
		items:      cards,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get cards")
}

type output struct {
	items      []models.Card
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
