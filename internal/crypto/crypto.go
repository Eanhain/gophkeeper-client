// Package crypto provides symmetric encryption helpers based on AES-256-GCM.
//
// The same functions are used for two purposes:
//   - encrypting the local cache file on disk
//   - encrypting HTTP request/response bodies between client and server
//
// Key derivation: SHA-256 of a passphrase → 32-byte AES key.
// Ciphertext format: 12-byte nonce || GCM-sealed data.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// DeriveKey produces a deterministic 32-byte AES-256 key from an
// arbitrary-length passphrase by taking its SHA-256 hash.
func DeriveKey(passphrase string) []byte {
	h := sha256.Sum256([]byte(passphrase))
	return h[:]
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// The key must be exactly 32 bytes (use DeriveKey).
// Returns nonce || ciphertext.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("rand nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt reverses Encrypt. It expects the ciphertext in the form
// nonce || GCM-sealed data. Returns an error if the key is wrong
// or the data has been tampered with (GCM authentication).
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, data, nil)
}

// EncryptString is a convenience wrapper that encrypts a UTF-8 string
// and returns the result as a base64-encoded string (safe for JSON).
func EncryptString(key []byte, plaintext string) (string, error) {
	encrypted, err := Encrypt(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString is the inverse of EncryptString: base64-decode then decrypt.
func DecryptString(key []byte, ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	decrypted, err := Decrypt(key, data)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
