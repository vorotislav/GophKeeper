package auth

import (
	"GophKeeper/internal/http/server/middlewares/auth/mocks"
	"GophKeeper/internal/token"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCheckAuth(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name        string
		prepareAuth func(auth *mocks.Authorizer)
		giveRequest func() *http.Request
		checkResult func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "excluded login uri",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing.com/login", http.NoBody)
				req.RequestURI = "login"
				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "excluded register uri",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing.com/register", http.NoBody)
				req.RequestURI = "register"
				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "without authorization header",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing.com/foo", http.NoBody)
				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, "Authorization field is empty\n", rr.Body.String())
			},
		},
		{
			name: "invalid authorization schema",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing.com/foo", http.NoBody)
				req.Header.Set("Authorization", "some-auth")
				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, "failed to check authorization token\n", rr.Body.String())
			},
		},
		{
			name: "parse error",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing.com/foo", http.NoBody)
				req.Header.Set("Authorization", "Bearer some-token")
				return req
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().ParseToken(mock.AnythingOfType("string")).
					Once().
					Return(token.Payload{}, errors.New("some error"))
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, "failed parse authorization token: some error\n", rr.Body.String())
			},
		},
		{
			name: "success",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing.com/foo", http.NoBody)
				req.Header.Set("Authorization", "Bearer some-token")
				return req
			},
			prepareAuth: func(auth *mocks.Authorizer) {
				auth.EXPECT().ParseToken(mock.AnythingOfType("string")).
					Once().
					Return(token.Payload{ID: 1}, nil)
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			})

			auth := mocks.NewAuthorizer(t)
			if tc.prepareAuth != nil {
				tc.prepareAuth(auth)
			}

			mw := CheckAuth(log, auth)(handler)
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, tc.giveRequest())

			tc.checkResult(t, rr)
		})
	}
}
