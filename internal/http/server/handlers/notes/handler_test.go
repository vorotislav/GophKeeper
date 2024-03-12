package notes

import (
	"GophKeeper/internal/http/server/handlers/notes/mocks"
	"GophKeeper/internal/models"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHandler(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	np := mocks.NewNoteProvider(t)
	require.NotNil(t, np)

	h := NewHandler(log, np)
	require.NotNil(t, h)
}

func TestHandler_NoteCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.NoteProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "invalid body",
			giveRequest: func() *http.Request {
				body := []byte(`some body`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note decode")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteCreate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "note":"text"}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteCreate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(fmt.Errorf("%w: %w", models.ErrInvalidInput, errors.New("some error")))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteCreate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, rr.Code)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			np := mocks.NewNoteProvider(t)
			require.NotNil(t, np)
			if tc.prepareProvider != nil {
				tc.prepareProvider(np)
			}

			h := &Handler{
				log:          log,
				noteProvider: np,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.NoteCreate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_NoteUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.NoteProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "invalid body",
			giveRequest: func() *http.Request {
				body := []byte(`some body`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note decode")
			},
		},
		{
			name: "provider not found error",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteUpdate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(models.ErrNotFound)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note update")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteUpdate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteUpdate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(fmt.Errorf("%w: %w", models.ErrInvalidInput, errors.New("some error")))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteUpdate(mock.Anything, mock.AnythingOfType("models.Note")).
					Once().
					Return(nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusAccepted, rr.Code)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			np := mocks.NewNoteProvider(t)
			require.NotNil(t, np)
			if tc.prepareProvider != nil {
				tc.prepareProvider(np)
			}

			h := &Handler{
				log:          log,
				noteProvider: np,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.NoteUpdate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_NoteDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.NoteProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "invalid path",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("noteID", "invalid")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note delete")
			},
		},
		{
			name: "not found",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("noteID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note delete")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("noteID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed note delete")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("noteID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().NoteDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNoContent, rr.Code)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			np := mocks.NewNoteProvider(t)
			require.NotNil(t, np)
			if tc.prepareProvider != nil {
				tc.prepareProvider(np)
			}

			h := &Handler{
				log:          log,
				noteProvider: np,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.NoteDelete(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_Notes(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.NoteProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "not found",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().Notes(mock.Anything).
					Once().Return(nil, fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get notes")
			},
		},
		{
			name: "not found 2",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().Notes(mock.Anything).
					Once().Return([]models.Note{}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get notes")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().Notes(mock.Anything).
					Once().Return(nil, errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get notes")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.NoteProvider) {
				provider.EXPECT().Notes(mock.Anything).
					Once().Return([]models.Note{
					{
						ID:        1,
						Title:     "title",
						Text:      "text",
						ExpiredAt: time.Time{},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.JSONEq(t, `[{"id":1, "title":"title", "note":"text", "created_at":"0001-01-01 00:00:00", "expired_at":"", "updated_at":"0001-01-01 00:00:00"}]`, rr.Body.String())
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			np := mocks.NewNoteProvider(t)
			require.NotNil(t, np)
			if tc.prepareProvider != nil {
				tc.prepareProvider(np)
			}

			h := &Handler{
				log:          log,
				noteProvider: np,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.Notes(rr, req)
			tc.checkResult(t, rr)
		})
	}
}
