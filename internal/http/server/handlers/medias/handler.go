package medias

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

type MediaProvider interface {
	MediaCreate(ctx context.Context, m models.Media) error
	MediaUpdate(ctx context.Context, m models.Media) error
	MediaDelete(ctx context.Context, id int) error
	Medias(ctx context.Context) ([]models.Media, error)
}

type Handler struct {
	log      *zap.Logger
	provider MediaProvider
}

func NewHandler(log *zap.Logger, mp MediaProvider) *Handler {
	return &Handler{
		log:      log.Named("media handler"),
		provider: mp,
	}
}

func (h *Handler) MediaCreate(w http.ResponseWriter, r *http.Request) {
	var (
		in      input
		expDate time.Time
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed media decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed media decode", err.Error()))

		return
	}

	if in.ExpiredAt != "" {
		expDate, err = time.Parse(inputTimeFormLong, in.ExpiredAt)
		if err != nil {
			h.log.Error("failed media exp date parse", zap.Error(err))

			responder.JSON(w, httpErr.NewInvalidInput("failed media exp date pars", err.Error()))

			return
		}
	}

	err = h.provider.MediaCreate(r.Context(), models.Media{
		Title:     in.Title,
		Body:      in.Media,
		MediaType: in.MediaType,
		Note:      in.Note,
		ExpiredAt: expDate,
	})

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

func (h *Handler) MediaUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		in      input
		expDate time.Time
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed media decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed media decode", err.Error()))

		return
	}

	if in.ExpiredAt != "" {
		expDate, err = time.Parse(inputTimeFormLong, in.ExpiredAt)
		if err != nil {
			h.log.Error("failed media exp date parse", zap.Error(err))

			responder.JSON(w, httpErr.NewInvalidInput("failed media exp date pars", err.Error()))

			return
		}
	}

	err = h.provider.MediaUpdate(r.Context(), models.Media{
		ID:        in.ID,
		Title:     in.Title,
		Body:      in.Media,
		MediaType: in.MediaType,
		Note:      in.Note,
		ExpiredAt: expDate,
	})

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
		if errors.Is(err, models.ErrNotFound) {
			responder.JSON(w, httpErr.NewNotFoundError("failed get medias"))

			return
		}
	}

	items := make([]item, 0, len(notes))
	for _, p := range notes {
		i := item{
			ID:        p.ID,
			Title:     p.Title,
			Media:     p.Body,
			MediaType: p.MediaType,
			Note:      p.Note,
			ExpiredAt: p.ExpiredAt.Format(inputTimeFormLong),
			CreatedAt: p.CreatedAt.Format(inputTimeFormLong),
			UpdatedAt: p.UpdatedAt.Format(inputTimeFormLong),
		}

		items = append(items, i)
	}

	responder.JSON(w, &output{
		items:      items,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get medias")
}

type input struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Media     []byte `json:"media"`
	MediaType string `json:"media_type"`
	Note      string `json:"note"`
	ExpiredAt string `json:"expired_at"`
}

type item struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Media     []byte `json:"media"`
	MediaType string `json:"media_type"`
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
