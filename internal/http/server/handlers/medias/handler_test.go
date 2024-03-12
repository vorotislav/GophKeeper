package medias

import (
	"GophKeeper/internal/http/server/handlers/medias/mocks"
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

	mp := mocks.NewMediaProvider(t)
	require.NotNil(t, mp)

	h := NewHandler(log, mp)
	require.NotNil(t, h)
}

func TestHandler_MediaCreate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.MediaProvider)
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
				assert.Contains(t, rr.Body.String(), "failed media decode")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaCreate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaCreate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(fmt.Errorf("%w: %w", models.ErrInvalidInput, errors.New("some error")))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media create")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaCreate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
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

			mp := mocks.NewMediaProvider(t)
			require.NotNil(t, mp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(mp)
			}

			h := &Handler{
				log:      log,
				provider: mp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.MediaCreate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_MediaUpdate(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.MediaProvider)
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
				assert.Contains(t, rr.Body.String(), "failed media decode")
			},
		},
		{
			name: "provider not found error",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaUpdate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(models.ErrNotFound)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media update")
			},
		},
		{
			name: "provider internal error",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaUpdate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "provider input error",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaUpdate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(fmt.Errorf("%w: %w", models.ErrInvalidInput, errors.New("some error")))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
				req, _ := http.NewRequest(http.MethodPut, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media update")
				assert.Contains(t, rr.Body.String(), "some error")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaUpdate(mock.Anything, mock.AnythingOfType("models.Media")).
					Once().
					Return(nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"title":"foo", "media":[104,101,108,108,111]}`)
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

			mp := mocks.NewMediaProvider(t)
			require.NotNil(t, mp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(mp)
			}

			h := &Handler{
				log:      log,
				provider: mp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.MediaUpdate(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_MediaDelete(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.MediaProvider)
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
				assert.Contains(t, rr.Body.String(), "failed media delete")
			},
		},
		{
			name: "not found",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("mediaID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media delete")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("mediaID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaDelete(mock.Anything, mock.AnythingOfType("int")).
					Once().Return(errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed media delete")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("mediaID", "2")

				req, _ := http.NewRequest(http.MethodDelete, "foo/bar/2", http.NoBody)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().MediaDelete(mock.Anything, mock.AnythingOfType("int")).
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

			mp := mocks.NewMediaProvider(t)
			require.NotNil(t, mp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(mp)
			}

			h := &Handler{
				log:      log,
				provider: mp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.MediaDelete(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_Medias(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.MediaProvider)
		giveRequest     func() *http.Request
		checkResult     func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "not found",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().Medias(mock.Anything).
					Once().Return(nil, fmt.Errorf("%w: some error", models.ErrNotFound))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get medias")
			},
		},
		{
			name: "not found 2",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().Medias(mock.Anything).
					Once().Return([]models.Media{}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get medias")
			},
		},
		{
			name: "internal error",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().Medias(mock.Anything).
					Once().Return(nil, errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed get medias")
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "foo/bar", http.NoBody)

				return req
			},
			prepareProvider: func(provider *mocks.MediaProvider) {
				provider.EXPECT().Medias(mock.Anything).
					Once().Return([]models.Media{
					{
						ID:        1,
						Title:     "title",
						Body:      []byte{},
						MediaType: "type",
						Note:      "",
						ExpiredAt: time.Time{},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.JSONEq(t, `[{"id":1, "title":"title", "media":"", "media_type":"type", "note":"", "expired_at":"0001-01-01 00:00:00", "created_at":"0001-01-01 00:00:00", "updated_at":"0001-01-01 00:00:00"}]`, rr.Body.String())
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mp := mocks.NewMediaProvider(t)
			require.NotNil(t, mp)
			if tc.prepareProvider != nil {
				tc.prepareProvider(mp)
			}

			h := &Handler{
				log:      log,
				provider: mp,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.Medias(rr, req)
			tc.checkResult(t, rr)
		})
	}
}
