// Package usecase implements the client-side business logic.
//
// It coordinates between the HTTP client (server communication) and
// the local encrypted cache, providing a single entry point for the TUI layer.
// Dependencies are expressed as interfaces so that the usecase can be
// fully tested with mocks — no real server or disk access required.
package usecase

import (
	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
)

// HTTPClient abstracts all REST API calls to the GophKeeper server.
// The concrete implementation lives in the clientconn package and uses
// Fiber's HTTP client with AES-256-GCM body encryption.
type HTTPClient interface {
	Register(login, password string) error
	Login(login, password string) (string, error)
	GetAllSecrets(token string) (*response.AllSecrets, error)
	PostLoginPassword(token string, lp request.LoginPassword) error
	PostTextSecret(token string, ts request.TextSecret) error
	PostBinarySecret(token string, bs request.BinarySecret) error
	PostCardSecret(token string, cs request.CardSecret) error
	DeleteLoginPassword(token, login string) error
	DeleteTextSecret(token, title string) error
	DeleteBinarySecret(token, filename string) error
	DeleteCardSecret(token, cardholder string) error
}

// SecretCache abstracts the encrypted local cache.
// Get returns nil when the cache is empty or stale.
// Write operations (Add*, Delete*) call Reset() to invalidate the cache
// so that the next read fetches fresh data from the server.
type SecretCache interface {
	Get() *response.AllSecrets
	Set(secrets *response.AllSecrets) error
	Reset()
}
