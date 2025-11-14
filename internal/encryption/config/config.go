package config

type EncryptionConfig struct {
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"` // path to the crypto key for the encryption.
}
