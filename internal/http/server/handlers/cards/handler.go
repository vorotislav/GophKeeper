package cards

import (
	httpErr "GophKeeper/internal/http/handlererrors"
	"GophKeeper/internal/http/responder"
	"GophKeeper/internal/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	inputTimeFormLong = "2006-01-02 15:04:05"
)

type CardProvider interface {
	CardCreate(ctx context.Context, c models.Card) error
	CardUpdate(ctx context.Context, c models.Card) error
	CardDelete(ctx context.Context, id int) error
	Cards(ctx context.Context) ([]models.Card, error)
}

type Handler struct {
	log          *zap.Logger
	cardProvider CardProvider
}

func NewHandler(log *zap.Logger, cp CardProvider) *Handler {
	return &Handler{
		log:          log.Named("cards handler"),
		cardProvider: cp,
	}
}

func (h *Handler) CardCreate(w http.ResponseWriter, r *http.Request) {
	var (
		in  input
		err error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed card decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed card decode", err.Error()))

		return
	}

	err = h.cardProvider.CardCreate(r.Context(), models.Card{
		Name:     in.Name,
		Number:   in.Card,
		CVC:      strconv.Itoa(in.CVV),
		ExpMonth: in.ExpMonth,
		ExpYear:  in.ExpYear,
	})

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

func (h *Handler) CardUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		in  input
		err error
	)

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.log.Error("failed card decode", zap.Error(err))

		responder.JSON(w, httpErr.NewInvalidInput("failed card decode", err.Error()))

		return
	}

	err = h.cardProvider.CardUpdate(r.Context(), models.Card{
		ID:       in.ID,
		Name:     in.Name,
		Number:   in.Card,
		CVC:      strconv.Itoa(in.CVV),
		ExpMonth: in.ExpMonth,
		ExpYear:  in.ExpYear,
	})

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
		if errors.Is(err, models.ErrNotFound) {
			responder.JSON(w, httpErr.NewNotFoundError("failed get cards"))

			return
		}
	}

	items := make([]item, 0, len(cards))
	for _, c := range cards {

		cvc, err := strconv.Atoi(c.CVC)
		if err != nil {
			responder.JSON(w, httpErr.NewInternalError("failed get cards", err.Error()))

			return
		}

		i := item{
			ID:        c.ID,
			Name:      c.Name,
			Card:      c.Number,
			ExpMonth:  c.ExpMonth,
			ExpYear:   c.ExpYear,
			CVV:       cvc,
			CreatedAt: c.CreatedAt.Format(inputTimeFormLong),
			UpdatedAt: c.UpdatedAt.Format(inputTimeFormLong),
		}

		items = append(items, i)
	}

	responder.JSON(w, &output{
		items:      items,
		statusCode: http.StatusOK,
	})

	h.log.Debug("success get cards")
}

type input struct {
	ID       int    `json:"ID"`
	Name     string `json:"name"`
	Card     string `json:"card"`
	ExpMonth int    `json:"expired_month_at"`
	ExpYear  int    `json:"expired_year_at"`
	CVV      int    `json:"cvv"`
}

type item struct {
	ID        int    `json:"ID"`
	Name      string `json:"name"`
	Card      string `json:"card"`
	ExpMonth  int    `json:"expired_month_at"`
	ExpYear   int    `json:"expired_year_at"`
	CVV       int    `json:"cvv"`
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
