// Package storage provides an encrypted SQLite-based cache for user secrets.
//
// The cache stores secrets in a local SQLite database file. Sensitive fields
// (passwords, PAN, text bodies, binary data) are encrypted with AES-256-GCM
// using a local passphrase (CRYPTO_KEY). This allows the client to serve
// read requests locally when the server is unavailable (offline mode).
//
// Invalidation strategy: every write/delete operation calls Reset(), which
// clears all cached data. The next read triggers a fresh fetch from the server.
package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"sync"

	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/crypto"

	_ "modernc.org/sqlite"
)

const dbFile = ".gophkeeper_cache.db"

// Cache is a thread-safe encrypted SQLite secret store.
// It implements the usecase.SecretCache interface.
type Cache struct {
	mu       sync.RWMutex
	key      []byte // AES-256 key derived from the passphrase
	db       *sql.DB
	secrets  *response.AllSecrets // in-memory cache
	wrongKey bool                 // true if data exists but decryption failed
}

// NewCache creates a new Cache with the given passphrase.
// Call Load() after creation to open the database and restore cached data.
func NewCache(cryptoKey string) *Cache {
	return &Cache{
		key: crypto.DeriveKey(cryptoKey),
	}
}

// Load opens (or creates) the SQLite database and restores any
// previously cached secrets into memory. Returns nil if the database
// does not exist — the cache simply stays empty.
func (c *Cache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	c.db = db

	// Create the cache table if it doesn't exist.
	_, err = c.db.Exec(`CREATE TABLE IF NOT EXISTS cache (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		data BLOB NOT NULL
	)`)
	if err != nil {
		return err
	}

	// Try to load existing data.
	var encData []byte
	err = c.db.QueryRow("SELECT data FROM cache WHERE id = 1").Scan(&encData)
	if err != nil {
		// No cached data — that's fine.
		return nil
	}

	plain, err := crypto.Decrypt(c.key, encData)
	if err != nil {
		// Data exists but decryption failed — wrong key.
		c.wrongKey = true
		return nil
	}

	var secrets response.AllSecrets
	if err := json.Unmarshal(plain, &secrets); err != nil {
		return nil
	}

	c.secrets = &secrets
	return nil
}

// WrongKey returns true if the database contains cached data but
// decryption failed during Load() — meaning the CRYPTO_KEY is incorrect.
func (c *Cache) WrongKey() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.wrongKey
}

// Get returns the cached secrets or nil if the cache is empty.
func (c *Cache) Get() *response.AllSecrets {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.secrets
}

// Set updates both the in-memory cache and the SQLite database.
func (c *Cache) Set(secrets *response.AllSecrets) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets = secrets
	return c.saveToDB()
}

// Reset clears the in-memory cache and removes all data from the SQLite database.
// This forces the next read to fetch fresh data from the server.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets = nil
	if c.db != nil {
		c.db.Exec("DELETE FROM cache") //nolint:errcheck
	}
}

// Close closes the underlying SQLite database connection.
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Remove deletes the SQLite database file from disk.
func (c *Cache) Remove() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		c.db.Close()
		c.db = nil
	}
	c.secrets = nil
	os.Remove(dbFile)
}

// saveToDB serialises the secrets to JSON, encrypts the result and writes
// it to the SQLite database using UPSERT.
func (c *Cache) saveToDB() error {
	if c.secrets == nil || c.db == nil {
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

	_, err = c.db.Exec(
		"INSERT INTO cache (id, data) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET data = ?",
		encrypted, encrypted,
	)
	return err
}
