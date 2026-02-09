package storage

import (
	"os"
	"testing"

	"github.com/Eanhain/gophkeeper-client/contracts/response"
)

func testSecrets() *response.AllSecrets {
	return &response.AllSecrets{
		LoginPassword: []response.LoginPassword{
			{Login: "admin", Password: "pass", Label: "work"},
		},
		TextSecret: []response.TextSecret{
			{Title: "note", Body: "hello"},
		},
	}
}

func TestCacheSetGet(t *testing.T) {
	os.Remove(cacheFile)
	defer os.Remove(cacheFile)

	c := NewCache("test-key")

	if c.Get() != nil {
		t.Fatal("empty cache must return nil")
	}

	secrets := testSecrets()
	if err := c.Set(secrets); err != nil {
		t.Fatalf("set: %v", err)
	}

	got := c.Get()
	if got == nil {
		t.Fatal("expected cached data")
	}
	if len(got.LoginPassword) != 1 || got.LoginPassword[0].Login != "admin" {
		t.Fatal("cached data mismatch")
	}
}

func TestCacheReset(t *testing.T) {
	os.Remove(cacheFile)
	defer os.Remove(cacheFile)

	c := NewCache("test-key")
	c.Set(testSecrets())
	c.Reset()

	if c.Get() != nil {
		t.Fatal("reset cache must return nil")
	}
	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Fatal("cache file should be deleted after reset")
	}
}

func TestCacheLoadFromDisk(t *testing.T) {
	os.Remove(cacheFile)
	defer os.Remove(cacheFile)

	c1 := NewCache("test-key")
	c1.Set(testSecrets())

	c2 := NewCache("test-key")
	if err := c2.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}

	got := c2.Get()
	if got == nil {
		t.Fatal("expected loaded data")
	}
	if got.LoginPassword[0].Login != "admin" {
		t.Fatal("loaded data mismatch")
	}
}

func TestCacheLoadWrongKey(t *testing.T) {
	os.Remove(cacheFile)
	defer os.Remove(cacheFile)

	c1 := NewCache("key1")
	c1.Set(testSecrets())

	c2 := NewCache("key2")
	err := c2.Load()
	if err == nil {
		t.Fatal("expected error loading with wrong key")
	}
}

func TestCacheLoadNoFile(t *testing.T) {
	os.Remove(cacheFile)
	c := NewCache("key")
	if err := c.Load(); err != nil {
		t.Fatalf("load missing file should not error: %v", err)
	}
	if c.Get() != nil {
		t.Fatal("expected nil for missing file")
	}
}

func TestCacheSetNil(t *testing.T) {
	os.Remove(cacheFile)
	defer os.Remove(cacheFile)

	c := NewCache("key")
	if err := c.Set(nil); err != nil {
		t.Fatalf("set nil: %v", err)
	}
	if c.Get() != nil {
		t.Fatal("expected nil after set nil")
	}
}
