package config

type EncryptionConfig struct {
	CryptoKey string `env:"CRYPTO_KEY"` // path to the crypto key for the encryption.
}
