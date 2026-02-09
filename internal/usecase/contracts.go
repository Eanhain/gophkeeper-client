package usecase

import (
	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
)

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

type SecretCache interface {
	Get() *response.AllSecrets
	Set(secrets *response.AllSecrets) error
	Reset()
}
