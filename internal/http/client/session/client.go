package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	ch "GophKeeper/internal/http"
	"GophKeeper/internal/http/client"
	"GophKeeper/internal/models"

	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
)

type systems interface {
	MachineInfo() (models.Machine, error)
}

type sessionStore interface {
	SaveSession(ses models.Session)
}

type Client struct {
	dc           *http.Client
	log          *zap.Logger
	serverURL    string
	systems      systems
	sessionStore sessionStore
}

func NewClient(
	log *zap.Logger,
	systems systems,
	sessionStore sessionStore,
	serverAddress string,
	httpTransport *http.Transport,
) *Client {
	c := &Client{
		dc: &http.Client{
			Timeout:   client.HTTPClientTimeout,
			Transport: httpTransport,
		},
		log:          log.Named("user client"),
		serverURL:    fmt.Sprintf("https://%s", serverAddress),
		systems:      systems,
		sessionStore: sessionStore,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.serverURL))

	return c
}

func (c *Client) Login(user models.User) (models.Session, error) {
	c.log.Debug("new request for user login")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	machine, err := c.systems.MachineInfo()
	if err != nil {
		c.log.Error("get machine info", zap.Error(err))
	}

	raw, err := json.Marshal(models.UserMachine{
		User:    user,
		Machine: machine,
	})
	if err != nil {
		c.log.Error("marshal to login", zap.Error(err))

		return models.Session{}, fmt.Errorf("login user: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+ch.LoginPath,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("login request prepare", zap.Error(err))

		return models.Session{}, fmt.Errorf("login user: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	var (
		body       []byte
		statusCode int
	)

	err = retry.Do(
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
		return models.Session{}, fmt.Errorf("cannot do request: %w", err)
	}

	c.log.Debug("user login", zap.Int("status code", statusCode))

	if statusCode == http.StatusNotFound {
		return models.Session{}, models.ErrNotFound
	}

	if statusCode == http.StatusBadRequest {
		return models.Session{}, models.ErrInvalidPassword
	}

	if statusCode != http.StatusOK {
		return models.Session{}, fmt.Errorf("login user")
	}

	s := models.Session{}

	err = json.Unmarshal(body, &s)
	if err != nil {
		return models.Session{}, fmt.Errorf("cannot unmarshal: %w", err)
	}

	c.sessionStore.SaveSession(s)

	return s, nil
}

func (c *Client) Register(user models.User) (models.Session, error) {
	c.log.Debug("new request for user register")

	ctx, cancel := context.WithTimeout(context.Background(), client.HTTPRequestTimeout)
	defer cancel()

	machine, err := c.systems.MachineInfo()
	if err != nil {
		c.log.Error("get machine info", zap.Error(err))
	}

	raw, err := json.Marshal(models.UserMachine{
		User:    user,
		Machine: machine,
	})
	if err != nil {
		c.log.Error("marshal to register", zap.Error(err))

		return models.Session{}, fmt.Errorf("register user: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+ch.RegisterPath,
		bytes.NewBuffer(raw))
	if err != nil {
		c.log.Error("register request prepare", zap.Error(err))

		return models.Session{}, fmt.Errorf("register user: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	var (
		body       []byte
		statusCode int
	)

	err = retry.Do(
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
		return models.Session{}, fmt.Errorf("cannot do request: %w", err)
	}

	if statusCode == http.StatusBadRequest {
		return models.Session{}, models.ErrInvalidInput
	}

	if statusCode != http.StatusOK {
		return models.Session{}, fmt.Errorf("register user")
	}

	s := models.Session{}

	err = json.Unmarshal(body, &s)
	if err != nil {
		return models.Session{}, fmt.Errorf("cannot unmarshal: %w", err)
	}

	c.log.Debug("Session", zap.Int64("id", s.ID), zap.String("access token", s.AccessToken))

	c.sessionStore.SaveSession(s)

	return s, nil
}
