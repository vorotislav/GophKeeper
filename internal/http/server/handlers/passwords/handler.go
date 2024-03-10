package passwords

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	inputTimeFormLong = "2006-01-02 15:04:05"
)

type PasswordProvider interface {
	PasswordCreate(ctx context.Context, p models.Password) error
	PasswordUpdate(ctx context.Context, p models.Password) error
	PasswordDelete(ctx context.Context, id int) error
	Passwords(ctx context.Context) ([]models.Password, error)
}

type Handler struct {
	log              *zap.Logger
	passwordProvider PasswordProvider
}

func NewHandler(log *zap.Logger, pp PasswordProvider) *Handler {
	return &Handler{
		log:              log.Named("passwords handler"),
		passwordProvider: pp,
	}
}

func (h *Handler) PasswordCreate(w http.ResponseWriter, r *http.Request) {
	var (
		in      input
		expDate time.Time
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed password decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed password decode", err.Error()))

		return
	}

	if in.ExpiredAt != "" {
		expDate, err = time.Parse(inputTimeFormLong, in.ExpiredAt)
		if err != nil {
			h.log.Error("failed password exp date parse", zap.Error(err))

			responder.JSON(w, httpErr.NewInvalidInput("failed password exp date parse", err.Error()))

			return
		}
	}

	err = h.passwordProvider.PasswordCreate(r.Context(), models.Password{
		Title:          in.Title,
		Login:          in.Login,
		Password:       in.Password,
		URL:            in.URL,
		Note:           in.Notes,
		ExpirationDate: expDate,
	})

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

func (h *Handler) PasswordUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		in      input
		expDate time.Time
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed password decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed password decode", err.Error()))

		return
	}

	if in.ExpiredAt != "" {
		expDate, err = time.Parse(inputTimeFormLong, in.ExpiredAt)
		if err != nil {
			h.log.Error("failed exp date password parse", zap.Error(err))

			responder.JSON(w, httpErr.NewInvalidInput("failed password exp date parse", err.Error()))

			return
		}
	}

	err = h.passwordProvider.PasswordUpdate(r.Context(), models.Password{
		ID:             in.ID,
		Title:          in.Title,
		Login:          in.Login,
		Password:       in.Password,
		URL:            in.URL,
		Note:           in.Notes,
		ExpirationDate: expDate,
	})

	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			h.log.Info("passwords not found")

			responder.JSON(w, httpErr.NewNotFoundError("passwords not found"))

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

			responder.JSON(w, httpErr.NewNotFoundError("password not found"))

			return
		}

		responder.JSON(w, httpErr.NewInternalError("failed password delete", err.Error()))

		return
	}

	setHeaders(w, http.StatusNoContent)
}

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
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, fmt.Sprintf("failed get passwords: %s", err.Error()), http.StatusNotFound)

			return
		}
	}

	items := make([]item, 0, len(passs))
	for _, p := range passs {
		var expDate string
		if !p.ExpirationDate.IsZero() {
			expDate = p.ExpirationDate.Format(inputTimeFormLong)
		}
		i := item{
			ID:        p.ID,
			Title:     p.Title,
			Login:     p.Login,
			Password:  p.Password,
			URL:       p.URL,
			Notes:     p.Note,
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

	h.log.Debug("success get passwords")
}

type input struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	URL       string `json:"url"`
	Notes     string `json:"notes"`
	ExpiredAt string `json:"expired_at"`
}

type item struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	URL       string `json:"url"`
	Notes     string `json:"notes"`
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
