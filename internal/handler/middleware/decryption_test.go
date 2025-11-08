package handler

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/encryption"
	"github.com/devize-ed/yapracproj-metrics.git/internal/logger"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

// setupTestKeys creates temporary RSA key files for testing.
func setupTestKeys(t *testing.T) (privKeyPath, pubKeyPath string) {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	privKeyPath = filepath.Join(tmpDir, "private_key.pem")
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	require.NoError(t, err)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	err = os.WriteFile(privKeyPath, privKeyPEM, 0600)
	require.NoError(t, err)

	pubKeyPath = filepath.Join(tmpDir, "public_key.pem")
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})
	err = os.WriteFile(pubKeyPath, pubKeyPEM, 0644)
	require.NoError(t, err)

	return privKeyPath, pubKeyPath
}

func TestDecryptionMiddleware(t *testing.T) {
	logger, err := logger.Initialize("debug")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	requestBody := `{
		"id":"LastGC",
		"type":"gauge"
	}`

	successBody := `{
		"id":"LastGC",
		"type":"gauge",
		"value":1744184459
	}`

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successBody))
	})

	privKeyPath, pubKeyPath := setupTestKeys(t)

	encryptor, err := encryption.NewEncryptor(pubKeyPath)
	require.NoError(t, err)

	tests := []struct {
		name        string
		privKeyPath string
		encryptBody bool
		encType     string
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "request_with_valid_encryption",
			privKeyPath: privKeyPath,
			encryptBody: true,
			encType:     "rsa",
			wantStatus:  http.StatusOK,
			wantBody:    successBody,
		},
		{
			name:        "missing_encryption_header",
			privKeyPath: privKeyPath,
			encryptBody: false,
			encType:     "",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "unsupported encryption type\n",
		},
		{
			name:        "unsupported_encryption_type",
			privKeyPath: privKeyPath,
			encryptBody: false,
			encType:     "aes",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "unsupported encryption type\n",
		},
		{
			name:        "empty_key_path",
			privKeyPath: "",
			encryptBody: true,
			encType:     "rsa",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "Encryption required\n",
		},
		{
			name:        "invalid_encrypted_data",
			privKeyPath: privKeyPath,
			encryptBody: false,
			encType:     "rsa",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "Error decrypting request body\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			middleware, err := DecryptionMiddleware(test.privKeyPath, logger)
			require.NoError(t, err)

			router := chi.NewRouter()
			router.Use(middleware)
			router.Post("/", successHandler)

			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			var body []byte
			if test.encryptBody {
				encryptedBody, err := encryptor.Encrypt([]byte(requestBody))
				require.NoError(t, err)
				body = encryptedBody
			} else if test.name == "invalid_encrypted_data" {
				body = []byte("invalid encrypted data")
			} else {
				body = []byte(requestBody)
			}

			req := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(body)

			if test.encType != "" {
				req.SetHeader("X-Encryption", test.encType)
			}

			resp, err := req.Post(srv.URL + "/")
			require.NoError(t, err)
			require.Equal(t, test.wantStatus, resp.StatusCode())
			require.Equal(t, test.wantBody, string(resp.Body()))
		})
	}
}

