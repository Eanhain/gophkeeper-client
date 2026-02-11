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
// The concrete implementation lives in the clientconn package.
type HTTPClient interface {
	Register(login, password string) (string, error)
	Login(login, password string) (string, error)
	GetAllSecrets(token string) (*response.AllSecrets, error)
	GetLoginPasswords(token string) ([]response.LoginPassword, error)
	GetTextSecrets(token string) ([]response.TextSecret, error)
	GetBinarySecrets(token string) ([]response.BinarySecret, error)
	GetCardSecrets(token string) ([]response.CardSecret, error)
	PostLoginPassword(token string, lp request.LoginPassword) error
	PostTextSecret(token string, ts request.TextSecret) error
	PostBinarySecret(token string, bs request.BinarySecret) error
	PostCardSecret(token string, cs request.CardSecret) error
	DeleteLoginPassword(token, login string) error
	DeleteTextSecret(token, title string) error
	DeleteBinarySecret(token, filename string) error
	DeleteCardSecret(token, cardholder string) error
}

// SecretCache abstracts the encrypted local cache (SQLite-backed).
// Get returns nil when the cache is empty or stale.
// Write operations (Add*, Delete*) call Reset() to invalidate the cache
// so that the next read fetches fresh data from the server.
type SecretCache interface {
	Get() *response.AllSecrets
	Set(secrets *response.AllSecrets) error
	Reset()
	Load() error
	Close() error
	WrongKey() bool
}
