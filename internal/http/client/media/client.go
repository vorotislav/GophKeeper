package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"io"
	"net/http"
	"strconv"
	"time"

	"GophKeeper/internal/models"

	"go.uber.org/zap"
)

const (
	httpClientTimeout = time.Millisecond * 500000
)

const (
	mediaPath = "/v1/medias"
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
		log:          log.Named("media client"),
		serverURL:    fmt.Sprintf("https://%s", serverAddress),
		sessionStore: ss,
	}

	log.Debug("Client for gophkeeper server", zap.String("url", c.serverURL))

	return c
}

func (c *Client) CreateMedia(m models.Media) error {
	c.log.Debug("new request for create media")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	var expDate string
	if !m.ExpiredAt.IsZero() {
		expDate = m.ExpiredAt.Format(inputTimeFormLong)
	}

	o := output{
		Title:     m.Title,
		Media:     m.Body,
		MediaType: m.MediaType,
		Note:      m.Note,
		ExpiredAt: expDate,
	}

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to create media", zap.Error(err))

		return fmt.Errorf("create media: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.serverURL+mediaPath,
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

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	var expDate string
	if !m.ExpiredAt.IsZero() {
		expDate = m.ExpiredAt.Format(inputTimeFormLong)
	}

	o := output{
		ID:        m.ID,
		Title:     m.Title,
		Media:     m.Body,
		MediaType: m.MediaType,
		Note:      m.Note,
		ExpiredAt: expDate,
	}

	raw, err := json.Marshal(o)
	if err != nil {
		c.log.Error("marshal to update media", zap.Error(err))

		return fmt.Errorf("update media: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		c.serverURL+mediaPath,
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

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.serverURL+mediaPath,
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

	items := make([]item, 0)
	if err := json.Unmarshal(body, &items); err != nil {
		c.log.Error("notes unmarshal", zap.Error(err))

		return nil, fmt.Errorf("notes unmarshal: %w", err)
	}

	medias := make([]models.Media, 0, len(items))

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

		media := models.Media{
			ID:        i.ID,
			Title:     i.Title,
			Body:      i.Media,
			MediaType: i.MediaType,
			Note:      i.Note,
			ExpiredAt: expDate,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		medias = append(medias, media)
	}

	return medias, nil
}

func (c *Client) DeleteMedia(id int) error {
	c.log.Debug("new request for delete media")

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	strID := strconv.Itoa(id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.serverURL+mediaPath+"/"+strID,
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

type output struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Media     []byte `json:"media"`
	MediaType string `json:"media_type"`
	Note      string `json:"note"`
	ExpiredAt string `json:"expired_at"`
}

type item struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Media     []byte `json:"media"`
	MediaType string `json:"media_type"`
	Note      string `json:"note"`
	ExpiredAt string `json:"expired_at"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
