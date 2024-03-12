// Package apptype описывает middleware для проверки наличия в заголовке Content-Type.
package apptype

import (
	"net/http"

	"go.uber.org/zap"
)

// ApplicationType создаёт middleware для проверки заголовка.
func ApplicationType(log *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" || r.Method == "DELETE" {
				h.ServeHTTP(w, r)

				return
			}

			if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
				log.Error("Failed to processing request", zap.String("unknown Content-Type", contentType))

				http.Error(w, "unknown Content-Type", http.StatusBadRequest)

				return
			}

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(ch)
	}
}
