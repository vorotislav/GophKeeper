package cards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"GophKeeper/internal/models"

	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
)

const (
	httpClientTimeout = time.Millisecond * 500000
)

const (
	cardsPath = "/v1/cards"
)

const (
	inputTimeFormLong = "2006-01-02 15:04:05"
)

type sessionStore interface {
	GetSession() models.Session
}

type Client struct {
	dc           *http.Client
	log          *zap.Logger
	serverURL    string
	sessionStore sessionStore
}

func NewClient(
	log *zap.Logger,
	ss sessionStore,
	serverAddress string,
	httpTransport *http.Transport,
) *Client {
	c := &Client{
		dc: &http.Client{
			Timeout:   httpClientTimeout,
			Transport: httpTransport,
		},
		log:          log.Named("cards client"),
		serverURL:    fmt.Sprintf("https://%s", serverAddress),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.serverURL))

	return c
}

func (c *Client) CreateCard(card models.Card) error {
	c.log.Debug("new request for create card")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	cvv, err := strconv.Atoi(card.CVC)
	if err != nil {
		c.log.Error("create card", zap.Error(err))

		return fmt.Errorf("create card: %w", err)
	}

	o := output{
		Name:     card.Name,
		Card:     card.Number,
		ExpMonth: card.ExpMonth,
		ExpYear:  card.ExpYear,
		CVV:      cvv,
	}

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to create card", zap.Error(err))

		return fmt.Errorf("create card: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+cardsPath,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("create card request prepare", zap.Error(err))

		return fmt.Errorf("create card: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("create card", zap.Error(err))

		return fmt.Errorf("create card: %w", err)
	}

	c.log.Debug("create card", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("create card")
	}

	return nil
}

func (c *Client) UpdateCard(card models.Card) error {
	c.log.Debug("new request for update card")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	cvv, err := strconv.Atoi(card.CVC)
	if err != nil {
		c.log.Error("update card", zap.Error(err))

		return fmt.Errorf("update card: %w", err)
	}

	o := output{
		ID:       card.ID,
		Name:     card.Name,
		Card:     card.Number,
		ExpMonth: card.ExpMonth,
		ExpYear:  card.ExpYear,
		CVV:      cvv,
	}

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to update card", zap.Error(err))

		return fmt.Errorf("update card: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.serverURL+cardsPath,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("update card request prepare", zap.Error(err))

		return fmt.Errorf("update card: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("update card", zap.Error(err))

		return fmt.Errorf("update card: %w", err)
	}

	c.log.Debug("update card", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusAccepted {
		return fmt.Errorf("update card")
	}

	return nil
}

func (c *Client) Cards() ([]models.Card, error) {
	c.log.Debug("new request for get cards")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.serverURL+cardsPath,
		http.NoBody)
	if err != nil {
		c.log.Error("get cards request prepare", zap.Error(err))

		return nil, fmt.Errorf("get cards: %w", err)
	}

	valueAuth := fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken)
	c.log.Debug("authorization set", zap.String("value", valueAuth))

	req.Header.Set("Authorization", valueAuth)

	body, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("get cards", zap.Error(err))

		return nil, fmt.Errorf("get cards: %w", err)
	}

	c.log.Debug("get cards", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return nil, models.ErrInvalidInput
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get cards")
	}

	items := make([]item, 0)
	if err := json.Unmarshal(body, &items); err != nil {
		c.log.Error("unmarshal cards", zap.Error(err))

		return nil, fmt.Errorf("unmarshal cards: %w", err)
	}

	cards := make([]models.Card, 0, len(items))
	for _, i := range items {
		createdAt, err := time.Parse(inputTimeFormLong, i.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("cannot parsing time: %w", err)
		}

		updatedAt, err := time.Parse(inputTimeFormLong, i.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("cannot parsing time: %w", err)
		}

		card := models.Card{
			ID:        i.ID,
			Name:      i.Name,
			Number:    i.Card,
			CVC:       strconv.Itoa(i.CVV),
			ExpMonth:  i.ExpMonth,
			ExpYear:   i.ExpYear,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		cards = append(cards, card)
	}

	return cards, nil
}

func (c *Client) DeleteCard(id int) error {
	c.log.Debug("new request for delete card")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.serverURL+cardsPath+"/"+strID,
		http.NoBody)
	if err != nil {
		c.log.Error("delete card request prepare", zap.Error(err))

		return fmt.Errorf("delete card: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("delete card", zap.Error(err))

		return fmt.Errorf("delete card: %w", err)
	}

	c.log.Debug("delete card", zap.Int("status code", statusCode))

	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusNoContent {
		return fmt.Errorf("delete card")
	}

	return nil
}

func (c *Client) do(ctx context.Context, req *http.Request) ([]byte, int, error) {
	var (
		body       []byte
		statusCode int
	)

	err := retry.Do(
		func() error {
			resp, err := c.dc.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = io.ReadAll(resp.Body)
			statusCode = resp.StatusCode

			if err != nil || resp.StatusCode >= http.StatusInternalServerError {
				return err
			}

			return nil
		},
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
		retry.Attempts(2),
		retry.Context(ctx))

	if err != nil {
		c.log.Error("cannot do request: %w", zap.Error(err))

		return nil, 0, fmt.Errorf("cannot do request: %w", err)
	}

	return body, statusCode, nil
}

type output struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Card     string `json:"card"`
	ExpMonth int    `json:"expired_month_at"`
	ExpYear  int    `json:"expired_year_at"`
	CVV      int    `json:"cvv"`
}

type item struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Card      string `json:"card"`
	ExpMonth  int    `json:"expired_month_at"`
	ExpYear   int    `json:"expired_year_at"`
	CVV       int    `json:"cvv"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
