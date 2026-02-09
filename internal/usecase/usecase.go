package usecase

import (
	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
	clientconn "github.com/Eanhain/gophkeeper-client/internal/clientConn"
	"github.com/Eanhain/gophkeeper-client/internal/storage"
)

type UseCase struct {
	client *clientconn.Client
	cache  *storage.Cache
	token  string
}

func New(client *clientconn.Client, cache *storage.Cache) *UseCase {
	return &UseCase{client: client, cache: cache}
}

func (uc *UseCase) SetToken(token string) {
	uc.token = token
}

func (uc *UseCase) Login(login, password string) (string, error) {
	return uc.client.Login(login, password)
}

func (uc *UseCase) Register(login, password string) (string, error) {
	if err := uc.client.Register(login, password); err != nil {
		return "", err
	}
	return uc.client.Login(login, password)
}

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

func (uc *UseCase) AddLoginPassword(lp request.LoginPassword) error {
	if err := uc.client.PostLoginPassword(uc.token, lp); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) AddTextSecret(ts request.TextSecret) error {
	if err := uc.client.PostTextSecret(uc.token, ts); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) AddBinarySecret(bs request.BinarySecret) error {
	if err := uc.client.PostBinarySecret(uc.token, bs); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) AddCardSecret(cs request.CardSecret) error {
	if err := uc.client.PostCardSecret(uc.token, cs); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) DeleteLoginPassword(login string) error {
	if err := uc.client.DeleteLoginPassword(uc.token, login); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) DeleteTextSecret(title string) error {
	if err := uc.client.DeleteTextSecret(uc.token, title); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) DeleteBinarySecret(filename string) error {
	if err := uc.client.DeleteBinarySecret(uc.token, filename); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) DeleteCardSecret(cardholder string) error {
	if err := uc.client.DeleteCardSecret(uc.token, cardholder); err != nil {
		return err
	}
	uc.cache.Reset()
	return nil
}

func (uc *UseCase) ResetCache() {
	uc.cache.Reset()
}
