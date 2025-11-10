// Package encryption provides RSA-OAEP helpers for encrypting and decrypting data.
//
// The package loads PEM-encoded RSA keys from disk and exposes two types:
//   - Encryptor: uses an RSA public key to encrypt data with OAEP and SHA-256.
//   - Decryptor: uses an RSA private key to decrypt data with OAEP and SHA-256.
//
// Public keys must be encoded as PKIX (x509.ParsePKIXPublicKey) in a PEM block.
// Private keys must be encoded as PKCS#8 (x509.ParsePKCS8PrivateKey) in a PEM block.
//
// Example:
//
//	enc, _ := encryption.NewEncryptor("public.pem")
//	ciphertext, _ := enc.Encrypt([]byte("secret"))
//
//	dec, _ := encryption.NewDecryptor("private.pem")
//	plaintext, _ := dec.Decrypt(ciphertext)
package encryption




