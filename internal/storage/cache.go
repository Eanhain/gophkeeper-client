// Package storage provides an encrypted file-based cache for user secrets.
//
// The cache stores a JSON-serialised [response.AllSecrets] blob encrypted
// with AES-256-GCM into the file ".gophkeeper_cache.enc" in the working
// directory. This allows the client to serve read requests locally without
// contacting the server.
//
// Invalidation strategy: every write/delete operation calls Reset(), which
// removes the in-memory state and deletes the disk file. The next read
// triggers a fresh fetch from the server.
package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/crypto"
)

// cacheFile is the name of the encrypted cache file on disk.
const cacheFile = ".gophkeeper_cache.enc"

// Cache is a thread-safe encrypted secret store.
// It implements the usecase.SecretCache interface.
type Cache struct {
	mu      sync.RWMutex
	key     []byte               // AES-256 key derived from the passphrase
	secrets *response.AllSecrets  // nil means "empty / stale"
}

// NewCache creates a new Cache with the given passphrase.
// Call Load() after creation to restore previously cached data from disk.
func NewCache(cryptoKey string) *Cache {
	return &Cache{
		key: crypto.DeriveKey(cryptoKey),
	}
}

// Load reads the encrypted cache file from disk and populates
// the in-memory state. Returns nil (no error) if the file does not exist —
// the cache simply stays empty. Returns an error if the file exists but
// cannot be decrypted (e.g. wrong key).
func (c *Cache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	plain, err := crypto.Decrypt(c.key, data)
	if err != nil {
		return err
	}

	var secrets response.AllSecrets
	if err := json.Unmarshal(plain, &secrets); err != nil {
		return err
	}

	c.secrets = &secrets
	return nil
}

// Get returns the cached secrets or nil if the cache is empty.
func (c *Cache) Get() *response.AllSecrets {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.secrets
}

// Set updates both the in-memory cache and the encrypted disk file.
func (c *Cache) Set(secrets *response.AllSecrets) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets = secrets
	return c.saveToDisk()
}

// Reset clears the in-memory cache and removes the disk file.
// This forces the next GetAllSecrets call to fetch fresh data from the server.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets = nil
	os.Remove(cacheFile)
}

// saveToDisk serialises the secrets to JSON, encrypts the result and writes
// it to disk with mode 0600 (owner read/write only).
func (c *Cache) saveToDisk() error {
	if c.secrets == nil {
		return nil
	}

	data, err := json.Marshal(c.secrets)
	if err != nil {
		return err
	}

	encrypted, err := crypto.Encrypt(c.key, data)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, encrypted, 0600)
}
