package decrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"
)

func DecryptMiddleware(log *zap.Logger, privateKeyPath string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := func(w http.ResponseWriter, r *http.Request) {
			privateKeyPEM, err := os.ReadFile(privateKeyPath)
			if err != nil {
				log.Debug("read private key path", zap.Error(err))

				h.ServeHTTP(w, r)

				return
			}

			block, _ := pem.Decode(privateKeyPEM)
			if block == nil {
				log.Debug("decode private public key")

				h.ServeHTTP(w, r)

				return
			}

			privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				log.Debug("parse private key", zap.Error(err))

				h.ServeHTTP(w, r)

				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = r.Body.Close()
			if err != nil {
				log.Error("http body close", zap.Error(err))
			}

			decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, body)
			if err != nil {
				log.Debug("decrypted body", zap.Error(err))

				http.Error(w, fmt.Sprintf("decrypted body: %s", err.Error()), http.StatusBadRequest)

				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decrypted))

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(ch)
	}
}
