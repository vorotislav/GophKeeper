package users

import (
	"GophKeeper/internal/http/server/handlers/users/mocks"
	"GophKeeper/internal/models"
	"bytes"
	"errors"
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

	up := mocks.NewUserProvider(t)
	require.NotNil(t, up)

	h := NewHandler(log, up)
	require.NotNil(t, h)
}

func TestHandler_Login(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.UserProvider)
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
				assert.Contains(t, rr.Body.String(), "failed to login user decode")
			},
		},
		{
			name: "provider error",
			prepareProvider: func(provider *mocks.UserProvider) {
				provider.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.UserMachine")).
					Once().
					Return(models.Session{}, errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"user":{"login":"login", "password":"pass"}, "machine":{"ip_address":"0.0.0.0"}}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed to login user")
			},
		},
		{
			name: "provider error",
			prepareProvider: func(provider *mocks.UserProvider) {
				provider.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.UserMachine")).
					Once().
					Return(models.Session{}, models.ErrNotFound)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"user":{"login":"login", "password":"pass"}, "machine":{"ip_address":"0.0.0.0"}}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				assert.Contains(t, rr.Body.String(), "user not found")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.UserProvider) {
				provider.EXPECT().UserLogin(mock.Anything, mock.AnythingOfType("models.UserMachine")).
					Once().
					Return(models.Session{
						ID:           1,
						UserID:       1,
						AccessToken:  "access",
						RefreshToken: "refresh",
						IPAddress:    "ip_address",
					}, nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"user":{"login":"login", "password":"pass"}, "machine":{"ip_address":"0.0.0.0"}}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.JSONEq(t, `{"id":1, "token":"access", "refresh_token":"refresh"}`, rr.Body.String())
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			up := mocks.NewUserProvider(t)
			require.NotNil(t, up)
			if tc.prepareProvider != nil {
				tc.prepareProvider(up)
			}

			h := &Handler{
				log:          log,
				userProvider: up,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.Login(rr, req)
			tc.checkResult(t, rr)
		})
	}
}

func TestHandler_Register(t *testing.T) {
	t.Parallel()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name            string
		prepareProvider func(provider *mocks.UserProvider)
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
				assert.Contains(t, rr.Body.String(), "failed to register user decode")
			},
		},
		{
			name: "provider error",
			prepareProvider: func(provider *mocks.UserProvider) {
				provider.EXPECT().UserCreate(mock.Anything, mock.AnythingOfType("models.UserMachine")).
					Once().
					Return(models.Session{}, errors.New("some error"))
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"user":{"login":"login", "password":"pass"}, "machine":{"ip_address":"0.0.0.0"}}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				assert.Contains(t, rr.Body.String(), "failed to register user")
			},
		},
		{
			name: "success",
			prepareProvider: func(provider *mocks.UserProvider) {
				provider.EXPECT().UserCreate(mock.Anything, mock.AnythingOfType("models.UserMachine")).
					Once().
					Return(models.Session{
						ID:           1,
						UserID:       1,
						AccessToken:  "access",
						RefreshToken: "refresh",
						IPAddress:    "ip_address",
					}, nil)
			},
			giveRequest: func() *http.Request {
				body := []byte(`{"user":{"login":"login", "password":"pass"}, "machine":{"ip_address":"0.0.0.0"}}`)
				req, _ := http.NewRequest(http.MethodPost, "foo/bar", bytes.NewBuffer(body))

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.JSONEq(t, `{"id":1, "token":"access", "refresh_token":"refresh"}`, rr.Body.String())
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			up := mocks.NewUserProvider(t)
			require.NotNil(t, up)
			if tc.prepareProvider != nil {
				tc.prepareProvider(up)
			}

			h := &Handler{
				log:          log,
				userProvider: up,
			}

			var (
				req = tc.giveRequest()
				rr  = httptest.NewRecorder()
			)

			h.Register(rr, req)
			tc.checkResult(t, rr)
		})
	}
}
