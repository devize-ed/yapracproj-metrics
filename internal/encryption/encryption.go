package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// EmptyCryptoKeyError is an error that is returned when a crypto key is empty.
var EmptyCryptoKeyError = errors.New("crypto key is empty")

// Encryptor is a struct that contains the public key for encryption.
type Encryptor struct {
	publicKey *rsa.PublicKey
}

// NewEncryptor creates a new Encryptor.
func NewEncryptor(path string) (*Encryptor, error) {
	// Read the key from the file.
	block, err := readKey(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}
	// Parse the public key from the PEM file and cast it to an RSA public key.
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}
	// Check if the public key is an RSA public key and cast it to an RSA public key.
	pubKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	// Return the Encryptor.
	return &Encryptor{publicKey: pubKey}, nil
}

// Encrypt encrypts the data using the public key.
func (e *Encryptor) Encrypt(data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, data, nil)
}

// Decryptor is a struct that contains the private key for decryption.
type Decryptor struct {
	privateKey *rsa.PrivateKey
}

// NewDecryptor creates a new Decryptor.
func NewDecryptor(path string) (*Decryptor, error) {
	// Read the key from the file.
	block, err := readKey(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}
	// Parse the private key from the PEM file
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}
	// Check if the private key is an RSA private key and cast it to an RSA private key.
	privKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	// Return the Decryptor.
	return &Decryptor{privateKey: privKey}, nil
}

// Decrypt decrypts the data using the private key.
func (d *Decryptor) Decrypt(data []byte) ([]byte, error) {
	// Decrypt the data.
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privateKey, data, nil)
}

// readKey reads the key from the file and returns a PEM block.
func readKey(path string) (*pem.Block, error) {
	// If the path is empty, return an error.
	if path == "" {
		return nil, EmptyCryptoKeyError
	}
	// Read the key from the file.
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file: %w", err)
	}
	// Decode the key from the PEM file.
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("no PEM block in key")
	}
	// Return the key.
	return block, nil
}
