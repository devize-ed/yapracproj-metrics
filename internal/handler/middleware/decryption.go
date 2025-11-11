package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/devize-ed/yapracproj-metrics.git/internal/encryption"
	"go.uber.org/zap"
)

// DecryptionMiddleware is a middleware that decrypts the request body.
func DecryptionMiddleware(privKeyPath string, logger *zap.SugaredLogger) (func(http.Handler) http.Handler, error) {
	// Create a new decryptor.
	decryptor, err := encryption.NewDecryptor(privKeyPath)
	if err != nil {
		if errors.Is(err, encryption.ErrEmptyCryptoKey) {
			logger.Debugf("Decryption key is empty, skipping decryption")
		} else {
			return nil, err
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the encryption type from the header.
			encType := r.Header.Get("X-Encryption")
			if encType != "rsa" {
				http.Error(w, "unsupported encryption type", http.StatusBadRequest)
				return
			}

			// If the decryptor is nil, check if encryption is required.
			if decryptor == nil {
				http.Error(w, "Encryption required", http.StatusBadRequest)
				return
			}

			// Read the body of the request.
			cipherBody, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Debugf("Error reading request body: %w", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			// Decrypt the request body.
			plain, err := decryptor.Decrypt(cipherBody)
			if err != nil {
				logger.Debugf("Error decrypting request body: %w", err)
				http.Error(w, "Error decrypting request body", http.StatusBadRequest)
				return
			}

			// Replace the body
			r.Body = io.NopCloser(bytes.NewReader(plain))
			r.ContentLength = int64(len(plain))

			// Serve the request.
			next.ServeHTTP(w, r)
		})
	}, nil
}
