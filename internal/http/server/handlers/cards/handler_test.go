package cards

import (
	"GophKeeper/internal/http/server/handlers/cards/mocks"
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

	cp := mocks.NewCardProvider(t)
	require.NotNil(t, cp)

	h := NewHandler(log, cp)
	require.NotNil(t, h)
}

func TestHandler_CardCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.CardProvider)
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
				assert.Contains(t, rr.Body.String(), "failed card decode")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardCreate(mock.Anything, mock.AnythingOfType("models.Card")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"name":"foo", "card":"card"}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed card create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardCreate(mock.Anything, mock.AnythingOfType("models.Card")).
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
				assert.Contains(t, rr.Body.String(), "failed card create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardCreate(mock.Anything, mock.AnythingOfType("models.Card")).
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

			cp := mocks.NewCardProvider(t)
			require.NotNil(t, cp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(cp)
			}

			h := &Handler{
				log:          log,
				cardProvider: cp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.CardCreate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_CardUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.CardProvider)
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
				assert.Contains(t, rr.Body.String(), "failed card decode")
			},
		},
		{
			name: "provider not found error",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardUpdate(mock.Anything, mock.AnythingOfType("models.Card")).
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
				assert.Contains(t, rr.Body.String(), "failed card update")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardUpdate(mock.Anything, mock.AnythingOfType("models.Card")).
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
				assert.Contains(t, rr.Body.String(), "failed card update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardUpdate(mock.Anything, mock.AnythingOfType("models.Card")).
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
				assert.Contains(t, rr.Body.String(), "failed card update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardUpdate(mock.Anything, mock.AnythingOfType("models.Card")).
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

			cp := mocks.NewCardProvider(t)
			require.NotNil(t, cp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(cp)
			}

			h := &Handler{
				log:          log,
				cardProvider: cp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.CardUpdate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_CardDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.CardProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "invalid path",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("cardID", "invalid")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed card delete")
			},
		},
		{
			name: "not found",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("cardID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed card delete")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("cardID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed card delete")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("cardID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().CardDelete(mock.Anything, mock.AnythingOfType("int")).
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

			cp := mocks.NewCardProvider(t)
			require.NotNil(t, cp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(cp)
			}

			h := &Handler{
				log:          log,
				cardProvider: cp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.CardDelete(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_Cards(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.CardProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "not found",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().Cards(mock.Anything).
					Once().Return(nil, fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get cards")
			},
		},
		{
			name: "not found 2",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().Cards(mock.Anything).
					Once().Return([]models.Card{}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get cards")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().Cards(mock.Anything).
					Once().Return(nil, errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get cards")
			},
		},
		{
			name: "invalid cvc",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().Cards(mock.Anything).
					Once().Return([]models.Card{
					{
						ID:       1,
						Name:     "name",
						Number:   "number",
						CVC:      "cvc",
						ExpMonth: 1,
						ExpYear:  2,
					},
				}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get cards")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.CardProvider) {
				provider.EXPECT().Cards(mock.Anything).
					Once().Return([]models.Card{
					{
						ID:       1,
						Name:     "name",
						Number:   "number",
						CVC:      "123",
						ExpMonth: 1,
						ExpYear:  2,
					},
				}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.JSONEq(t, `[{"id":1, "name":"name", "card":"number", "cvv":123, "expired_month_at":1, "expired_year_at":2, "created_at":"0001-01-01 00:00:00", "updated_at":"0001-01-01 00:00:00"}]`, rr.Body.String())
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cp := mocks.NewCardProvider(t)
			require.NotNil(t, cp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(cp)
			}

			h := &Handler{
				log:          log,
				cardProvider: cp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.Cards(rr, req)
			tc.checkResult(t, rr)
		})
	}
}
