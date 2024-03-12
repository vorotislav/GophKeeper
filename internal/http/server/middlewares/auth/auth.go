package auth

import (
	"GophKeeper/internal/token"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// authorizer описывает метод для парсинга и проверка входящего токена.
//
//go:generate go run github.com/vektra/mockery/v2@v2.24.0 --name=authorizer --exported --with-expecter=true
type authorizer interface {
	ParseToken(string) (token.Payload, error)
}

func CheckAuth(log *zap.Logger, auth authorizer) func(h http.Handler) http.Handler {
	const (
		headerAuthorization = "Authorization"
		authorizationSchema = "Bearer"
		excludedLoginURI    = "login"
		excludedRegisterURI = "register"
	)

	return func(h http.Handler) http.Handler {
		ch := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.RequestURI, excludedLoginURI) ||
				strings.Contains(r.RequestURI, excludedRegisterURI) {

				h.ServeHTTP(w, r)

				return
			}

			bearer := r.Header.Get(headerAuthorization)
			if bearer == "" {
				log.Error("failed to get authorization field")

				http.Error(w, "Authorization field is empty", http.StatusBadRequest)

				return
			}

			before, after, found := strings.Cut(bearer, " ")
			if !found && !strings.EqualFold(before, authorizationSchema) {
				log.Error("failed to check authorization token")

				http.Error(w, "failed to check authorization token", http.StatusBadRequest)

				return
			}

			tokenPaylod, err := auth.ParseToken(after)
			if err != nil {
				log.Error("failed parse authorization token", zap.Error(err))

				http.Error(w, fmt.Sprintf("failed parse authorization token: %s", err.Error()), http.StatusBadRequest)

				return
			}

			*r = *r.WithContext(token.ToContext(r.Context(), tokenPaylod))

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(ch)
	}
}
