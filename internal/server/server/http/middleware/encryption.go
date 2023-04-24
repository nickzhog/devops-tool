package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/nickzhog/devops-tool/pkg/encryption"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func RequestDecryptMiddleWare(key *rsa.PrivateKey, logger *logging.Logger) func(next http.Handler) http.Handler {
	fn := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if len(body) < 1 {
				next.ServeHTTP(w, r)
				return
			}
			decryptedBody, err := encryption.DecryptData(body, key)
			if err != nil {
				logger.Error(err)
				http.Error(w, "cant decrypt request data", http.StatusNotAcceptable)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
			next.ServeHTTP(w, r)
		})
	}
	return fn
}
