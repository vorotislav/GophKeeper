package password

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
	passwordsURL string
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
		log:          log.Named("passwords client"),
		passwordsURL: fmt.Sprintf("https://%s%s", serverAddress, ch.PasswordsPath),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.passwordsURL))

	return c
}

func (c *Client) CreatePassword(pass models.Password) error {
	c.log.Debug("new request for create password")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(pass)
	if err != nil {
		c.log.Error("marshal to create password", zap.Error(err))

		return fmt.Errorf("create password: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.passwordsURL,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("create password request prepare", zap.Error(err))

		return fmt.Errorf("create password: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("create password", zap.Error(err))

		return fmt.Errorf("create password: %w", err)
	}

	c.log.Debug("create password", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("create password")
	}

	return nil
}

func (c *Client) UpdatePassword(pass models.Password) error {
	c.log.Debug("new request for update password")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	raw, err := json.Marshal(pass)
	if err != nil {
		c.log.Error("marshal to update password", zap.Error(err))

		return fmt.Errorf("update password: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.passwordsURL,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("update password request prepare", zap.Error(err))

		return fmt.Errorf("update password: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("update password", zap.Error(err))

		return fmt.Errorf("update password: %w", err)
	}

	c.log.Debug("update password", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusAccepted {
		return fmt.Errorf("update password")
	}

	return nil
}

func (c *Client) Passwords() ([]models.Password, error) {
	c.log.Debug("new request for get passwords")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.passwordsURL,
		http.NoBody)
	if err != nil {
		c.log.Error("get passwords request prepare", zap.Error(err))

		return nil, fmt.Errorf("get passwords: %w", err)
	}

	valueAuth := fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken)
	c.log.Debug("authorization set", zap.String("value", valueAuth))

	req.Header.Set("Authorization", valueAuth)

	body, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("get passwords", zap.Error(err))

		return nil, fmt.Errorf("get passwords: %w", err)
	}

	c.log.Debug("get passwords", zap.Int("status code", statusCode))

	if statusCode == http.StatusBadRequest {
		return nil, models.ErrInvalidInput
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get passwords")
	}

	passwords := make([]models.Password, 0)
	err = json.Unmarshal(body, &passwords)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal: %w", err)
	}

	return passwords, nil
}

func (c *Client) DeletePassword(id int) error {
	c.log.Debug("new request for delete password")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.passwordsURL+"/"+strID,
		http.NoBody)
	if err != nil {
		c.log.Error("delete password request prepare", zap.Error(err))

		return fmt.Errorf("delete password: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.sessionStore.GetSession().AccessToken))

	_, statusCode, err := c.do(ctx, req)
	if err != nil {
		c.log.Error("delete password", zap.Error(err))

		return fmt.Errorf("delete password: %w", err)
	}

	c.log.Debug("delete password", zap.Int("status code", statusCode))

	if err != nil {
		return fmt.Errorf("cannot do request: %w", err)
	}

	if statusCode == http.StatusBadRequest {
		return models.ErrInvalidInput
	}

	if statusCode != http.StatusNoContent {
		return fmt.Errorf("delete password")
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
