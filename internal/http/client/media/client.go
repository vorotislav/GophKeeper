package media

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
	mediaURL     string
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
		log:          log.Named("media client"),
		mediaURL:     fmt.Sprintf("https://%s%s", serverAddress, ch.MediaPath),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.mediaURL))

	return c
}

func (c *Client) CreateMedia(m models.Media) error {
	c.log.Debug("new request for create media")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(m)
	if err != nil {
		c.log.Error("marshal to create media", zap.Error(err))

		return fmt.Errorf("create media: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.mediaURL,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("create media request prepare", zap.Error(err))

		return fmt.Errorf("create media: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("create media", zap.Error(err))

		return fmt.Errorf("create media: %w", err)
	}

	c.log.Debug("create media", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("create media")
	}

	return nil
}

func (c *Client) UpdateMedia(m models.Media) error {
	c.log.Debug("new request for update media")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(m)
	if err != nil {
		c.log.Error("marshal to update media", zap.Error(err))

		return fmt.Errorf("update media: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.mediaURL,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("update media request prepare", zap.Error(err))

		return fmt.Errorf("create media: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("update media", zap.Error(err))

		return fmt.Errorf("update media: %w", err)
	}

	c.log.Debug("update media", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusAccepted {
		return fmt.Errorf("update media")
	}

	return nil
}

func (c *Client) Medias() ([]models.Media, error) {
	c.log.Debug("new request for get notes")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.mediaURL,
		http.NoBody)
	if err != nil {
		c.log.Error("get notes request prepare", zap.Error(err))

		return nil, fmt.Errorf("get notes: %w", err)
	}

	valueAuth := fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken)
	c.log.Debug("authorization set", zap.String("value", valueAuth))

	req.Header.Set("Authorization", valueAuth)

	body, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("get notes", zap.Error(err))

		return nil, fmt.Errorf("get notes: %w", err)
	}

	c.log.Debug("get notes", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return nil, models.ErrInvalidInput
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get notes")
	}

	medias := make([]models.Media, 0)
	if err := json.Unmarshal(body, &medias); err != nil {
		c.log.Error("notes unmarshal", zap.Error(err))

		return nil, fmt.Errorf("notes unmarshal: %w", err)
	}

	return medias, nil
}

func (c *Client) DeleteMedia(id int) error {
	c.log.Debug("new request for delete media")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.mediaURL+"/"+strID,
		http.NoBody)
	if err != nil {
		c.log.Error("delete media request prepare", zap.Error(err))

		return fmt.Errorf("delete media: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("delete media", zap.Error(err))

		return fmt.Errorf("delete media: %w", err)
	}

	c.log.Debug("delete media", zap.Int("status code", statusCode))

	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusNoContent {
		return fmt.Errorf("delete media")
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
