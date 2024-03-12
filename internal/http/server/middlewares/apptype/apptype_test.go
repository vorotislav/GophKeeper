package apptype

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestApplicationType(t *testing.T) {
	const errText = "unknown Content-Type\n"

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	require.NotNil(t, log)

	cases := []struct {
		name        string
		checkResult func(t *testing.T, rr *httptest.ResponseRecorder)
		giveRequest func() *http.Request
	}{
		{
			name: "get",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "some url", http.NoBody)

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "delete",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodDelete, "some url", http.NoBody)

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "post without app type",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "some url", http.NoBody)

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, errText, rr.Body.String())
			},
		},
		{
			name: "put without app type",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPut, "some url", http.NoBody)

				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				assert.Equal(t, errText, rr.Body.String())
			},
		},
		{
			name: "post with app type",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "some url", http.NoBody)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			checkResult: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "put with app type",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPut, "some url", http.NoBody)
				req.Header.Set("Content-Type", "application/json")
				return req
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

			mw := ApplicationType(log)(handler)
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, tc.giveRequest())

			tc.checkResult(t, rr)
		})
	}
}
