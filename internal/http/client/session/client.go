package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"GophKeeper/internal/models"

	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
)

const (
	httpClientTimeout = time.Millisecond * 500000
)

const (
	loginPath    = "/v1/users/login"
	registerPath = "/v1/users/register"
)

type systems interface {
	MachineInfo() models.Machine
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
			Timeout:   httpClientTimeout,
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

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	o := getOut(user, c.systems.MachineInfo())

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to login", zap.Error(err))

		return models.Session{}, fmt.Errorf("login user: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+loginPath,
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

	i := in{}

	err = json.Unmarshal(body, &i)
	if err != nil {
		return models.Session{}, fmt.Errorf("cannot unmarshal: %w", err)
	}

	s := models.Session{
		ID:           i.ID,
		AccessToken:  i.Token,
		RefreshToken: i.RefreshToken,
	}

	c.sessionStore.SaveSession(s)

	return s, nil
}

func (c *Client) Register(user models.User) (models.Session, error) {
	c.log.Debug("new request for user register")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	o := getOut(user, c.systems.MachineInfo())

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to register", zap.Error(err))

		return models.Session{}, fmt.Errorf("register user: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+registerPath,
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

	i := in{}

	err = json.Unmarshal(body, &i)
	if err != nil {
		return models.Session{}, fmt.Errorf("cannot unmarshal: %w", err)
	}

	c.log.Debug("Session", zap.Int64("id", i.ID), zap.String("access token", i.Token))

	return models.Session{
		ID:           i.ID,
		AccessToken:  i.Token,
		RefreshToken: i.RefreshToken,
	}, nil
}

func getOut(u models.User, m models.Machine) out {
	return out{
		User: user{
			Login:    u.Login,
			Password: u.Password,
		},
		Machine: machine{
			IPAddress:  m.IPAddress,
			MACAddress: m.MACAddress,
			PublicKey:  m.PublicKey,
		},
	}
}

type out struct {
	User    user    `json:"user"`
	Machine machine `json:"machine"`
}

type user struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type machine struct {
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
	PublicKey  string `json:"public_key"`
}

type in struct {
	ID           int64  `json:"id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
