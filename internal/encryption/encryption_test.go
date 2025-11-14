package encryption_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/devize-ed/yapracproj-metrics.git/internal/encryption"
	"github.com/stretchr/testify/assert"
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

func TestNewEncryptor(t *testing.T) {
	_, pubKeyPath := setupTestKeys(t)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid_public_key",
			path:    pubKeyPath,
			wantErr: false,
		},
		{
			name:    "empty_path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "invalid_path",
			path:    "testdata/invalid.pem",
			wantErr: true,
		},
		{
			name:    "invalid_key_format",
			path:    "/dev/null",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := encryption.NewEncryptor(tt.path)
			if tt.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestEncryptor(t *testing.T) {
	_, pubKeyPath := setupTestKeys(t)

	encryptor, err := encryption.NewEncryptor(pubKeyPath)
	require.NoError(t, err)
	require.NotNil(t, encryptor)

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "encrypt_valid_data",
			data:    []byte("test_data"),
			wantErr: false,
		},
		{
			name:    "encrypt_empty_data",
			data:    []byte(""),
			wantErr: false,
		},
		{
			name:    "encrypt_large_data",
			data:    make([]byte, 100),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := encryptor.Encrypt(tt.data)
			if tt.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, got)
				assert.NotEqual(t, tt.data, got)
				assert.Greater(t, len(got), 0)
			}
		})
	}
}

func TestNewDecryptor(t *testing.T) {
	privKeyPath, _ := setupTestKeys(t)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid_private_key",
			path:    privKeyPath,
			wantErr: false,
		},
		{
			name:    "empty_path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "invalid_path",
			path:    "testdata/invalid.pem",
			wantErr: true,
		},
		{
			name:    "invalid_key_format",
			path:    "/dev/null",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := encryption.NewDecryptor(tt.path)
			if tt.wantErr {
				assert.Error(t, gotErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, gotErr)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestDecryptor(t *testing.T) {
	privKeyPath, pubKeyPath := setupTestKeys(t)

	decryptor, err := encryption.NewDecryptor(privKeyPath)
	require.NoError(t, err)
	require.NotNil(t, decryptor)

	encryptor, err := encryption.NewEncryptor(pubKeyPath)
	require.NoError(t, err)
	require.NotNil(t, encryptor)

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "decrypt_valid_data",
			data:    []byte("test_data"),
			wantErr: false,
		},
		{
			name:    "decrypt_empty_data",
			data:    []byte(""),
			wantErr: false,
		},
		{
			name:    "decrypt_large_data",
			data:    make([]byte, 100),
			wantErr: false,
		},
		{
			name:    "decrypt_invalid_data",
			data:    []byte("invalid encrypted data"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "decrypt_invalid_data" {
				got, gotErr := decryptor.Decrypt(tt.data)
				assert.Error(t, gotErr)
				assert.Nil(t, got)
			} else {
				encrypted, err := encryptor.Encrypt(tt.data)
				require.NoError(t, err)
				require.NotNil(t, encrypted)

				got, gotErr := decryptor.Decrypt(encrypted)
				if tt.wantErr {
					assert.Error(t, gotErr)
					assert.Nil(t, got)
				} else {
					assert.NoError(t, gotErr)
					assert.Equal(t, tt.data, got)
				}
			}
		})
	}
}
