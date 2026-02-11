package usecase

import (
	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
)

// UseCase orchestrates client-side operations.
//
// Read path:  try server → cache result → return.
//
//	If server unavailable → return from cache (offline mode).
//
// Write path: send to server → invalidate cache (Reset).
//
//	If server unavailable → return error (writes require server).
//
// This strategy provides offline read access while keeping writes consistent.
type UseCase struct {
	client HTTPClient  // talks to the GophKeeper server
	cache  SecretCache // encrypted local cache (SQLite)
	token  string      // JWT token received after login/register
}

// New creates a UseCase with the given HTTP client and cache.
func New(client HTTPClient, cache SecretCache) *UseCase {
	return &UseCase{client: client, cache: cache}
}

// SetToken stores the JWT token that will be sent with every
// authenticated request to the server.
func (uc *UseCase) SetToken(token string) {
	uc.token = token
}

// Login authenticates against the server and returns a JWT token.
func (uc *UseCase) Login(login, password string) (string, error) {
	return uc.client.Login(login, password)
}

// Register creates a new account and returns a JWT token on success.
func (uc *UseCase) Register(login, password string) (string, error) {
	return uc.client.Register(login, password)
}

// GetCachedSecrets returns whatever is in the local cache without
// contacting the server. Returns nil if the cache is empty.
func (uc *UseCase) GetCachedSecrets() *response.AllSecrets {
	return uc.cache.Get()
}

// GetAllSecrets tries to fetch secrets from the server.
// On success, the result is cached locally. On network failure,
// the cached version is returned (offline mode).
func (uc *UseCase) GetAllSecrets() (*response.AllSecrets, error) {
	secrets, err := uc.client.GetAllSecrets(uc.token)
	if err == nil {
		_ = uc.cache.Set(secrets)
		return secrets, nil
	}

	// Server unavailable — fall back to cache (offline mode).
	if cached := uc.cache.Get(); cached != nil {
		return cached, nil
	}

	// No cache either — return the original error.
	return nil, err
}

// GetLoginPasswords fetches login-password secrets from the server.
// Falls back to the cached version on network failure.
func (uc *UseCase) GetLoginPasswords() ([]response.LoginPassword, error) {
	result, err := uc.client.GetLoginPasswords(uc.token)
	if err == nil {
		return result, nil
	}
	// Offline fallback.
	if cached := uc.cache.Get(); cached != nil {
		return cached.LoginPassword, nil
	}
	return nil, err
}

// GetTextSecrets fetches text secrets from the server.
// Falls back to the cached version on network failure.
func (uc *UseCase) GetTextSecrets() ([]response.TextSecret, error) {
	result, err := uc.client.GetTextSecrets(uc.token)
	if err == nil {
		return result, nil
	}
	if cached := uc.cache.Get(); cached != nil {
		return cached.TextSecret, nil
	}
	return nil, err
}

// GetBinarySecrets fetches binary secrets from the server.
// Falls back to the cached version on network failure.
func (uc *UseCase) GetBinarySecrets() ([]response.BinarySecret, error) {
	result, err := uc.client.GetBinarySecrets(uc.token)
	if err == nil {
		return result, nil
	}
	if cached := uc.cache.Get(); cached != nil {
		return cached.BinarySecret, nil
	}
	return nil, err
}

// GetCardSecrets fetches card secrets from the server.
// Falls back to the cached version on network failure.
func (uc *UseCase) GetCardSecrets() ([]response.CardSecret, error) {
	result, err := uc.client.GetCardSecrets(uc.token)
	if err == nil {
		return result, nil
	}
	if cached := uc.cache.Get(); cached != nil {
		return cached.CardSecret, nil
	}
	return nil, err
}

// AddLoginPassword sends a new login-password to the server
// and invalidates the local cache.
func (uc *UseCase) AddLoginPassword(lp request.LoginPassword) error {
	if err := uc.client.PostLoginPassword(uc.token, lp); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// AddTextSecret sends a new text secret and invalidates the cache.
func (uc *UseCase) AddTextSecret(ts request.TextSecret) error {
	if err := uc.client.PostTextSecret(uc.token, ts); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// AddBinarySecret sends a new binary secret and invalidates the cache.
func (uc *UseCase) AddBinarySecret(bs request.BinarySecret) error {
	if err := uc.client.PostBinarySecret(uc.token, bs); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// AddCardSecret sends a new card secret and invalidates the cache.
func (uc *UseCase) AddCardSecret(cs request.CardSecret) error {
	if err := uc.client.PostCardSecret(uc.token, cs); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// DeleteLoginPassword deletes a login-password by its login identifier.
func (uc *UseCase) DeleteLoginPassword(login string) error {
	if err := uc.client.DeleteLoginPassword(uc.token, login); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// DeleteTextSecret deletes a text secret by its title.
func (uc *UseCase) DeleteTextSecret(title string) error {
	if err := uc.client.DeleteTextSecret(uc.token, title); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// DeleteBinarySecret deletes a binary secret by filename.
func (uc *UseCase) DeleteBinarySecret(filename string) error {
	if err := uc.client.DeleteBinarySecret(uc.token, filename); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// DeleteCardSecret deletes a card secret by cardholder name.
func (uc *UseCase) DeleteCardSecret(cardholder string) error {
	if err := uc.client.DeleteCardSecret(uc.token, cardholder); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

// ResetCache clears the local encrypted cache.
// Useful when a secret was changed by another client.
func (uc *UseCase) ResetCache() {
	uc.cache.Reset()
}

// IsWrongKey returns true if the local cache has data but the CRYPTO_KEY
// is incorrect (decryption failed during Load).
func (uc *UseCase) IsWrongKey() bool {
	return uc.cache.WrongKey()
}
