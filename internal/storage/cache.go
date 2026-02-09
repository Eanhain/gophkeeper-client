package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/crypto"
)

const cacheFile = ".gophkeeper_cache.enc"

type Cache struct {
	mu      sync.RWMutex
	key     []byte
	secrets *response.AllSecrets
}

func NewCache(cryptoKey string) *Cache {
	return &Cache{
		key: crypto.DeriveKey(cryptoKey),
	}
}

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

func (c *Cache) Get() *response.AllSecrets {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.secrets
}

func (c *Cache) Set(secrets *response.AllSecrets) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets = secrets
	return c.saveToDisk()
}

func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.secrets = nil
	os.Remove(cacheFile)
}

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
