## Encryption package

The `internal/encryption` package provides functions to encrypt and decrypt data using RSA-OAEP with SHA-256.

- **Encryptor**: Loads an RSA public key (PEM, PKIX) and encrypts bytes.
- **Decryptor**: Loads an RSA private key (PEM, PKCS#8) and decrypts bytes.
