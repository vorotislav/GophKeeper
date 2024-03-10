package password

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
	passwordsPath = "/v1/passwords"
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
		log:          log.Named("passwords client"),
		serverURL:    fmt.Sprintf("https://%s", serverAddress),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.serverURL))

	return c
}

func (c *Client) CreatePassword(pass models.Password) error {
	c.log.Debug("new request for create password")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	var expDate string
	if !pass.ExpirationDate.IsZero() {
		expDate = pass.ExpirationDate.Format(inputTimeFormLong)
	}

	o := out{
		Title:     pass.Title,
		Login:     pass.Login,
		Password:  pass.Password,
		URL:       pass.URL,
		Notes:     pass.Note,
		ExpiredAt: expDate,
	}

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to create password", zap.Error(err))

		return fmt.Errorf("create password: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+passwordsPath,
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

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	var expDate string
	if !pass.ExpirationDate.IsZero() {
		expDate = pass.ExpirationDate.Format(inputTimeFormLong)
	}

	o := out{
		ID:        pass.ID,
		Title:     pass.Title,
		Login:     pass.Login,
		Password:  pass.Password,
		URL:       pass.URL,
		Notes:     pass.Note,
		ExpiredAt: expDate,
	}

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to update password", zap.Error(err))

		return fmt.Errorf("update password: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.serverURL+passwordsPath,
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

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.serverURL+passwordsPath,
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

	items := make([]item, 0)
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal: %w", err)
	}

	passwords := make([]models.Password, 0, len(items))
	for _, i := range items {
		createdAt, err := time.Parse(inputTimeFormLong, i.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("cannot parsing time: %w", err)
		}

		updatedAt, err := time.Parse(inputTimeFormLong, i.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("cannot parsing time: %w", err)
		}

		var expDate time.Time

		if i.ExpiredAt != "" {
			expDate, err = time.Parse(inputTimeFormLong, i.ExpiredAt)
			if err != nil {
				return nil, fmt.Errorf("cannot parsing time: %w", err)
			}
		}

		p := models.Password{
			ID:             i.ID,
			Title:          i.Title,
			Login:          i.Login,
			Password:       i.Password,
			URL:            i.URL,
			Note:           i.Notes,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			ExpirationDate: expDate,
		}

		passwords = append(passwords, p)
	}

	return passwords, nil
}

func (c *Client) DeletePassword(id int) error {
	c.log.Debug("new request for delete password")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.serverURL+passwordsPath+"/"+strID,
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

type out struct {
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
