package usecase

import (
	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
)

// UseCase orchestrates client-side operations.
//
// Read path:  cache hit → return cached data.
//
//	cache miss → fetch from server → store in cache → return.
//
// Write path: send to server → invalidate cache (Reset).
//
// This cache-first strategy minimises network traffic while the
// cache-invalidation-on-write approach keeps data consistent.
type UseCase struct {
	client HTTPClient  // talks to the GophKeeper server
	cache  SecretCache // encrypted local cache
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

// Register creates a new account and immediately logs in,
// returning a JWT token on success.
func (uc *UseCase) Register(login, password string) (string, error) {
	if err := uc.client.Register(login, password); err != nil {
		return "", err
	}
	return uc.client.Login(login, password)
}

// GetAllSecrets returns secrets from cache if available,
// otherwise fetches them from the server and populates the cache.
func (uc *UseCase) GetAllSecrets() (*response.AllSecrets, error) {
	if cached := uc.cache.Get(); cached != nil {
		return cached, nil
	}

	secrets, err := uc.client.GetAllSecrets(uc.token)
	if err != nil {
		return nil, err
	}

	uc.cache.Set(secrets)
	return secrets, nil
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
