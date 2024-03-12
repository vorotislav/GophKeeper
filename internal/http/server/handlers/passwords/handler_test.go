package passwords

import (
	"GophKeeper/internal/http/server/handlers/passwords/mocks"
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
)

func TestNewHandler(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	pp := mocks.NewPasswordProvider(t)
	require.NotNil(t, pp)

	h := NewHandler(log, pp)
	require.NotNil(t, h)
}

func TestHandler_PasswordCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.PasswordProvider)
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
				assert.Contains(t, rr.Body.String(), "failed password decode")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordCreate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordCreate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(fmt.Errorf("%w: %w", models.ErrInvalidInput, errors.New("some error")))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordCreate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
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

			pp := mocks.NewPasswordProvider(t)
			require.NotNil(t, pp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(pp)
			}

			h := &Handler{
				log:              log,
				passwordProvider: pp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.PasswordCreate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_PasswordUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.PasswordProvider)
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
				assert.Contains(t, rr.Body.String(), "failed password decode")
			},
		},
		{
			name: "provider not found error",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordUpdate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(models.ErrNotFound)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password update")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordUpdate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordUpdate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(fmt.Errorf("%w: %w", models.ErrInvalidInput, errors.New("some error")))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordUpdate(mock.Anything, mock.AnythingOfType("models.Password")).
					Once().
					Return(nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"login":"foo", "password":"pass"}`)
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

			pp := mocks.NewPasswordProvider(t)
			require.NotNil(t, pp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(pp)
			}

			h := &Handler{
				log:              log,
				passwordProvider: pp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.PasswordUpdate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_PasswordDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.PasswordProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "invalid path",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("passwordID", "invalid")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password delete")
			},
		},
		{
			name: "not found",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("passwordID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password delete")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("passwordID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed password delete")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("passwordID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().PasswordDelete(mock.Anything, mock.AnythingOfType("int")).
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

			pp := mocks.NewPasswordProvider(t)
			require.NotNil(t, pp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(pp)
			}

			h := &Handler{
				log:              log,
				passwordProvider: pp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.PasswordDelete(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_Passwords(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.PasswordProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "not found",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().Passwords(mock.Anything).
					Once().Return(nil, fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get passwords")
			},
		},
		{
			name: "not found 2",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().Passwords(mock.Anything).
					Once().Return([]models.Password{}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get passwords")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().Passwords(mock.Anything).
					Once().Return(nil, errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get passwords")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.PasswordProvider) {
				provider.EXPECT().Passwords(mock.Anything).
					Once().Return([]models.Password{
					{
						ID:       1,
						Title:    "title",
						Login:    "login",
						Password: "pass",
						URL:      "url",
						Note:     "note",
					},
				}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.JSONEq(t, `[{"id":1, "title":"title", "login":"login", "password":"pass", "url":"url", "notes":"note", "expired_at":"", "created_at":"0001-01-01 00:00:00", "updated_at":"0001-01-01 00:00:00"}]`, rr.Body.String())
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pp := mocks.NewPasswordProvider(t)
			require.NotNil(t, pp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(pp)
			}

			h := &Handler{
				log:              log,
				passwordProvider: pp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.Passwords(rr, req)
			tc.checkResult(t, rr)
		})
	}
}
