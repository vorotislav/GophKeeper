package cards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	ch "GophKeeper/internal/http"
	"GophKeeper/internal/http/client"
	"GophKeeper/internal/models"

	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
)

type sessionStore interface {
	GetSession() models.Session
}

type Client struct {
	dc           *http.Client
	log          *zap.Logger
	cardsURL     string
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
			Timeout:   client.HTTPClientTimeout,
			Transport: httpTransport,
		},
		log:          log.Named("cards client"),
		cardsURL:     fmt.Sprintf("https://%s%s", serverAddress, ch.CardsPath),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.cardsURL))

	return c
}

func (c *Client) CreateCard(card models.Card) error {
	c.log.Debug("new request for create card")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(card)
	if err != nil {
		c.log.Error("marshal to create card", zap.Error(err))

		return fmt.Errorf("create card: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.cardsURL,
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

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(card)
	if err != nil {
		c.log.Error("marshal to update card", zap.Error(err))

		return fmt.Errorf("update card: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.cardsURL,
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

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.cardsURL,
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

	cards := make([]models.Card, 0)
	if err := json.Unmarshal(body, &cards); err != nil {
		c.log.Error("unmarshal cards", zap.Error(err))

		return nil, fmt.Errorf("unmarshal cards: %w", err)
	}

	return cards, nil
}

func (c *Client) DeleteCard(id int) error {
	c.log.Debug("new request for delete card")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.cardsURL+"/"+strID,
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
