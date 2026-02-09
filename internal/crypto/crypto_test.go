package crypto

import (
	"testing"
)

func TestDeriveKey(t *testing.T) {
	key := DeriveKey("test-passphrase")
	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}
	key2 := DeriveKey("test-passphrase")
	if string(key) != string(key2) {
		t.Fatal("same passphrase must produce same key")
	}
	key3 := DeriveKey("other")
	if string(key) == string(key3) {
		t.Fatal("different passphrases must produce different keys")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := DeriveKey("secret")
	plain := []byte("hello world")

	encrypted, err := Encrypt(key, plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if string(encrypted) == string(plain) {
		t.Fatal("encrypted must differ from plaintext")
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(decrypted) != string(plain) {
		t.Fatalf("expected %q, got %q", plain, decrypted)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := DeriveKey("key1")
	key2 := DeriveKey("key2")

	encrypted, _ := Encrypt(key1, []byte("data"))
	_, err := Decrypt(key2, encrypted)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := DeriveKey("key")
	_, err := Decrypt(key, []byte("short"))
	if err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestEncryptStringDecryptString(t *testing.T) {
	key := DeriveKey("key")

	enc, err := EncryptString(key, "секрет")
	if err != nil {
		t.Fatalf("encrypt string: %v", err)
	}
	if enc == "секрет" {
		t.Fatal("encrypted string must differ")
	}

	dec, err := DecryptString(key, enc)
	if err != nil {
		t.Fatalf("decrypt string: %v", err)
	}
	if dec != "секрет" {
		t.Fatalf("expected секрет, got %s", dec)
	}
}

func TestDecryptStringInvalidBase64(t *testing.T) {
	key := DeriveKey("key")
	_, err := DecryptString(key, "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestEncryptEmpty(t *testing.T) {
	key := DeriveKey("key")
	enc, err := Encrypt(key, []byte{})
	if err != nil {
		t.Fatalf("encrypt empty: %v", err)
	}
	dec, err := Decrypt(key, enc)
	if err != nil {
		t.Fatalf("decrypt empty: %v", err)
	}
	if len(dec) != 0 {
		t.Fatal("expected empty plaintext")
	}
}

func TestEncryptBadKey(t *testing.T) {
	_, err := Encrypt([]byte("short"), []byte("data"))
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestDecryptBadKey(t *testing.T) {
	_, err := Decrypt([]byte("short"), []byte("some-ciphertext-that-is-long-enough-for-nonce"))
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestEncryptStringBadKey(t *testing.T) {
	_, err := EncryptString([]byte("short"), "data")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDecryptStringBadKey(t *testing.T) {
	key := DeriveKey("key")
	enc, _ := EncryptString(key, "hello")
	_, err := DecryptString([]byte("short"), enc)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEncryptDecryptLargeData(t *testing.T) {
	key := DeriveKey("key")
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	enc, err := Encrypt(key, data)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := Decrypt(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if len(dec) != len(data) {
		t.Fatalf("length mismatch: %d vs %d", len(dec), len(data))
	}
}
