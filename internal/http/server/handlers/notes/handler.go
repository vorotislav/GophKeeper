// Package notes реализует HTTP обработчик для /notes
package notes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"

	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"

	"go.uber.org/zap"
)

const (
	inputTimeFormLong = "2006-01-02 15:04:05"
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
		in      input
		expDate time.Time
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed note decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed note decode", err.Error()))

		return
	}

	if in.ExpiredAt != "" {
		expDate, err = time.Parse(inputTimeFormLong, in.ExpiredAt)
		if err != nil {
			h.log.Error("failed note exp date parse", zap.Error(err))

			responder.JSON(w, httpErr.NewInvalidInput("failed note exp date pars", err.Error()))

			return
		}
	}

	err = h.noteProvider.NoteCreate(r.Context(), models.Note{
		Title:     in.Title,
		Text:      in.Note,
		ExpiredAt: expDate,
	})

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
		in      input
		expDate time.Time
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed note decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed note decode", err.Error()))

		return
	}

	if in.ExpiredAt != "" {
		expDate, err = time.Parse(inputTimeFormLong, in.ExpiredAt)
		if err != nil {
			h.log.Error("failed note exp date parse", zap.Error(err))

			responder.JSON(w, httpErr.NewInvalidInput("failed note exp date pars", err.Error()))

			return
		}
	}

	err = h.noteProvider.NoteUpdate(r.Context(), models.Note{
		ID:        in.ID,
		Title:     in.Title,
		Text:      in.Note,
		ExpiredAt: expDate,
	})

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

	items := make([]item, 0, len(notes))
	for _, p := range notes {
		var expDate string
		if !p.ExpiredAt.IsZero() {
			expDate = p.ExpiredAt.String()
		}
		i := item{
			ID:        p.ID,
			Title:     p.Title,
			Note:      p.Text,
			ExpiredAt: expDate,
			CreatedAt: p.CreatedAt.Format(inputTimeFormLong),
			UpdatedAt: p.UpdatedAt.Format(inputTimeFormLong),
		}

		items = append(items, i)
	}

	responder.JSON(w, &output{
		items:      items,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get notes")
}

type input struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Note      string `json:"note"`
	ExpiredAt string `json:"expired_at"`
}

type item struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Note      string `json:"note"`
	ExpiredAt string `json:"expired_at"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type output struct {
	items      []item
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
