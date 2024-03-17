// Package notes реализует HTTP обработчик для /notes
package notes

import (
	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

// NoteProvider описывает методы для работы с заметками пользователей.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=NoteProvider --with-expecter=true
type NoteProvider interface {
	NoteCreate(ctx context.Context, n models.Note) error
	NoteUpdate(ctx context.Context, n models.Note) error
	NoteDelete(ctx context.Context, id int) error
	Notes(ctx context.Context) ([]models.Note, error)
}

// Handler реализует HTTP обработчик для /notes.
type Handler struct {
	log          *zap.Logger
	noteProvider NoteProvider
}

// NewHandler конструктор для Handler.
func NewHandler(log *zap.Logger, np NoteProvider) *Handler {
	return &Handler{
		log:          log.Named("notes handler"),
		noteProvider: np,
	}
}

// NoteCreate обрабатывает POST запрос на создание заметки.
func (h *Handler) NoteCreate(w http.ResponseWriter, r *http.Request) {
	var (
		note models.Note
		err  error
	)

	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		h.log.Error("failed note decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed note decode", err.Error()))

		return
	}

	err = h.noteProvider.NoteCreate(r.Context(), note)

	if err != nil {
		h.log.Error("failed note create", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed note create", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed note create", err.Error()))

		return
	}

	setHeaders(w, http.StatusCreated)

	h.log.Info("success note create")
}

// NoteUpdate обрабатывает PUT запрос на обновление заметки.
func (h *Handler) NoteUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		note models.Note
		err  error
	)

	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		h.log.Error("failed note decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed note decode", err.Error()))

		return
	}

	err = h.noteProvider.NoteUpdate(r.Context(), note)

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("note not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed note update"))

			return
		}

		h.log.Error("failed note update", zap.Error(err))

		if errors.Is(err, models.ErrInvalidInput) {
			responder.JSON(w, httpErr.NewInvalidInput("failed note update", err.Error()))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed note update", err.Error()))

		return
	}

	setHeaders(w, http.StatusAccepted)
	h.log.Debug("success update note")
}

// NoteDelete обрабатывает DELETE запрос на удаление заметки.
func (h *Handler) NoteDelete(w http.ResponseWriter, r *http.Request) {
	strNoteID := chi.URLParam(r, "noteID")
	noteID, err := strconv.Atoi(strNoteID)
	if err != nil {
		h.log.Error("failed get note id", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed note delete", err.Error()))

		return
	}

	err = h.noteProvider.NoteDelete(r.Context(), noteID)
	if err != nil {
		h.log.Error("failed note delete", zap.Error(err))

		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("note not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed note delete"))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed note delete", err.Error()))

		return
	}

	setHeaders(w, http.StatusNoContent)

	h.log.Debug("success delete notes")
}

// Notes обрабатывает GET запрос на получение заметок.
func (h *Handler) Notes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.noteProvider.Notes(r.Context())
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("notes not found")

			responder.JSON(w, httpErr.NewNotFoundError("failed get notes"))

			return
		}

		h.log.Error("failed get notes", zap.Error(err))

		responder.JSON(w, httpErr.NewInternalError("failed get notes", err.Error()))

		return
	}

	if len(notes) == 0 {
		responder.JSON(w, httpErr.NewNotFoundError("failed get notes"))

		return
	}

	responder.JSON(w, &output{
		items:      notes,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get notes")
}

type output struct {
	items      []models.Note
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
