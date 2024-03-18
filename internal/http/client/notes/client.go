package notes

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
	noteURL      string
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
		log:          log.Named("notes client"),
		noteURL:      fmt.Sprintf("https://%s%s", serverAddress, ch.NotesPath),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.noteURL))

	return c
}

func (c *Client) CreateNote(n models.Note) error {
	c.log.Debug("new request for create note")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(n)
	if err != nil {
		c.log.Error("marshal to create note", zap.Error(err))

		return fmt.Errorf("create note: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.noteURL,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("create note request prepare", zap.Error(err))

		return fmt.Errorf("create note: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("create note", zap.Error(err))

		return fmt.Errorf("create note: %w", err)
	}

	c.log.Debug("create note", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("create note")
	}

	return nil
}

func (c *Client) UpdateNote(n models.Note) error {
	c.log.Debug("new request for update note")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(n)
	if err != nil {
		c.log.Error("marshal to update note", zap.Error(err))

		return fmt.Errorf("update note: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.noteURL,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("update note request prepare", zap.Error(err))

		return fmt.Errorf("create note: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("update note", zap.Error(err))

		return fmt.Errorf("update note: %w", err)
	}

	c.log.Debug("update note", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusAccepted {
		return fmt.Errorf("update note")
	}

	return nil
}

func (c *Client) Notes() ([]models.Note, error) {
	c.log.Debug("new request for get notes")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.noteURL,
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

	notes := make([]models.Note, 0)
	if err := json.Unmarshal(body, &notes); err != nil {
		c.log.Error("notes unmarshal", zap.Error(err))

		return nil, fmt.Errorf("notes unmarshal: %w", err)
	}

	return notes, nil
}

func (c *Client) DeleteNote(id int) error {
	c.log.Debug("new request for delete note")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.noteURL+"/"+strID,
		http.NoBody)
	if err != nil {
		c.log.Error("delete note request prepare", zap.Error(err))

		return fmt.Errorf("delete note: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("delete note", zap.Error(err))

		return fmt.Errorf("delete note: %w", err)
	}

	c.log.Debug("delete note", zap.Int("status code", statusCode))

	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusNoContent {
		return fmt.Errorf("delete note")
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
