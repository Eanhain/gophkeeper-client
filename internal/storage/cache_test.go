package storage_test

import (
	"os"
	"testing"

	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/storage"
)

const testDBFile = ".gophkeeper_cache.db"

func cleanup() {
	os.Remove(testDBFile)
}

func TestCache_SetAndGet(t *testing.T) {
	defer cleanup()

	cache := storage.NewCache("test-passphrase")
	if err := cache.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	defer cache.Close()

	// Initially empty.
	if cache.Get() != nil {
		t.Fatal("expected nil cache initially")
	}

	secrets := &response.AllSecrets{
		LoginPassword: []response.LoginPassword{
			{Login: "admin", Password: "secret", Label: "test"},
		},
		TextSecret: []response.TextSecret{
			{Title: "note", Body: "hello world"},
		},
	}

	if err := cache.Set(secrets); err != nil {
		t.Fatalf("set: %v", err)
	}

	got := cache.Get()
	if got == nil {
		t.Fatal("expected non-nil cache")
	}
	if len(got.LoginPassword) != 1 || got.LoginPassword[0].Login != "admin" {
		t.Fatal("data mismatch")
	}
	if len(got.TextSecret) != 1 || got.TextSecret[0].Title != "note" {
		t.Fatal("text data mismatch")
	}
}

func TestCache_Reset(t *testing.T) {
	defer cleanup()

	cache := storage.NewCache("test-passphrase")
	if err := cache.Load(); err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	_ = cache.Set(&response.AllSecrets{
		LoginPassword: []response.LoginPassword{{Login: "a"}},
	})

	cache.Reset()
	if cache.Get() != nil {
		t.Fatal("expected nil after reset")
	}
}

func TestCache_PersistenceAcrossLoads(t *testing.T) {
	defer cleanup()

	// Write data with first cache instance.
	c1 := storage.NewCache("persist-key")
	if err := c1.Load(); err != nil {
		t.Fatal(err)
	}
	_ = c1.Set(&response.AllSecrets{
		CardSecret: []response.CardSecret{{Cardholder: "John", Pan: "1234"}},
	})
	_ = c1.Close()

	// Read with a new cache instance (same key).
	c2 := storage.NewCache("persist-key")
	if err := c2.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	defer c2.Close()

	got := c2.Get()
	if got == nil {
		t.Fatal("expected persisted data")
	}
	if len(got.CardSecret) != 1 || got.CardSecret[0].Cardholder != "John" {
		t.Fatal("data mismatch after reload")
	}
}

func TestCache_WrongKeyIgnored(t *testing.T) {
	defer cleanup()

	// Write with one key.
	c1 := storage.NewCache("key1")
	if err := c1.Load(); err != nil {
		t.Fatal(err)
	}
	_ = c1.Set(&response.AllSecrets{
		LoginPassword: []response.LoginPassword{{Login: "x"}},
	})
	_ = c1.Close()

	// Try to read with a different key — should silently ignore.
	c2 := storage.NewCache("key2")
	if err := c2.Load(); err != nil {
		t.Fatal(err)
	}
	defer c2.Close()

	if c2.Get() != nil {
		t.Fatal("expected nil when loading with wrong key")
	}
}
