// Package medias реализует HTTP обработчик для /medias
package medias

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

// MediaProvider описывает методы для работы с медиа пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=MediaProvider --with-expecter=true
type MediaProvider interface {
	MediaCreate(ctx context.Context, m models.Media) error
	MediaUpdate(ctx context.Context, m models.Media) error
	MediaDelete(ctx context.Context, id int) error
	Medias(ctx context.Context) ([]models.Media, error)
}

// Handler реализует HTTP обработчик для /cards.
type Handler struct {
	log      *zap.Logger
	provider MediaProvider
}

// NewHandler конструктор для Handler.
func NewHandler(log *zap.Logger, mp MediaProvider) *Handler {
	return &Handler{
		log:      log.Named("media handler"),
		provider: mp,
	}
}

// MediaCreate обрабатывает POST запрос на создание медиа.
func (h *Handler) MediaCreate(w http.ResponseWriter, r *http.Request) {
	var (
		media models.Media
		err   error
	)

	if err := json.NewDecoder(r.Body).Decode(&media); err != nil {
		h.log.Error("failed media decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed media decode", err.Error()))

		return
	}

	err = h.provider.MediaCreate(r.Context(), media)

	if err != nil {
		h.log.Error("failed media create", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed media create", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed media create", err.Error()))

		return
	}

	setHeaders(w, http.StatusCreated)

	h.log.Info("success media create")
}

// MediaUpdate обрабатывает PUT запрос на обновление медиа.
func (h *Handler) MediaUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		media models.Media
		err   error
	)

	if err := json.NewDecoder(r.Body).Decode(&media); err != nil {
		h.log.Error("failed media decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed media decode", err.Error()))

		return
	}

	err = h.provider.MediaUpdate(r.Context(), media)

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("media not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed media update"))

			return
		}

		h.log.Error("failed media update", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed media update", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed media update", err.Error()))

		return
	}

	setHeaders(w, http.StatusAccepted)

	h.log.Info("success media update")
}

// MediaDelete обрабатывает DELETE запрос на удаление медиа.
func (h *Handler) MediaDelete(w http.ResponseWriter, r *http.Request) {
	strMediaID := chi.URLParam(r, "mediaID")
	mediaID, err := strconv.Atoi(strMediaID)
	if err != nil {
		h.log.Error("failed get media id", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed media delete", err.Error()))

		return
	}

	err = h.provider.MediaDelete(r.Context(), mediaID)
	if err != nil {
		h.log.Error("failed media delete", zap.Error(err))

		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("media not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed media delete"))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed media delete", err.Error()))

		return
	}

	setHeaders(w, http.StatusNoContent)

	h.log.Info("success media delete")
}

// Medias обрабатывает GET запрос на получение медиа.
func (h *Handler) Medias(w http.ResponseWriter, r *http.Request) {
	notes, err := h.provider.Medias(r.Context())
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("medias not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed get medias"))

			return
		}

		h.log.Error("failed get medias", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed get medias", err.Error()))

		return
	}

	if len(notes) == 0 {
		responder.JSON(w, httpErr.NewNotFoundError("failed get medias"))

		return
	}

	responder.JSON(w, &output{
		items:      notes,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get medias")
}

type output struct {
	items      []models.Media
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
