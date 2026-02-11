package usecase_test

import (
	"errors"
	"testing"

	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/usecase"
)

// --- mock HTTP client ---

type mockClient struct {
	registerFn          func(login, password string) (string, error)
	loginFn             func(login, password string) (string, error)
	getAllSecretsFn     func(token string) (*response.AllSecrets, error)
	getLoginPasswordsFn func(token string) ([]response.LoginPassword, error)
	getTextSecretsFn    func(token string) ([]response.TextSecret, error)
	getBinarySecretsFn  func(token string) ([]response.BinarySecret, error)
	getCardSecretsFn    func(token string) ([]response.CardSecret, error)
	postLoginPasswordFn func(token string, lp request.LoginPassword) error
	postTextSecretFn    func(token string, ts request.TextSecret) error
	postBinarySecretFn  func(token string, bs request.BinarySecret) error
	postCardSecretFn    func(token string, cs request.CardSecret) error
	deleteLoginPwdFn    func(token, login string) error
	deleteTextFn        func(token, title string) error
	deleteBinaryFn      func(token, filename string) error
	deleteCardFn        func(token, cardholder string) error
}

func (m *mockClient) Register(login, password string) (string, error) {
	return m.registerFn(login, password)
}
func (m *mockClient) Login(login, password string) (string, error) {
	return m.loginFn(login, password)
}
func (m *mockClient) GetAllSecrets(token string) (*response.AllSecrets, error) {
	return m.getAllSecretsFn(token)
}
func (m *mockClient) GetLoginPasswords(token string) ([]response.LoginPassword, error) {
	return m.getLoginPasswordsFn(token)
}
func (m *mockClient) GetTextSecrets(token string) ([]response.TextSecret, error) {
	return m.getTextSecretsFn(token)
}
func (m *mockClient) GetBinarySecrets(token string) ([]response.BinarySecret, error) {
	return m.getBinarySecretsFn(token)
}
func (m *mockClient) GetCardSecrets(token string) ([]response.CardSecret, error) {
	return m.getCardSecretsFn(token)
}
func (m *mockClient) PostLoginPassword(token string, lp request.LoginPassword) error {
	return m.postLoginPasswordFn(token, lp)
}
func (m *mockClient) PostTextSecret(token string, ts request.TextSecret) error {
	return m.postTextSecretFn(token, ts)
}
func (m *mockClient) PostBinarySecret(token string, bs request.BinarySecret) error {
	return m.postBinarySecretFn(token, bs)
}
func (m *mockClient) PostCardSecret(token string, cs request.CardSecret) error {
	return m.postCardSecretFn(token, cs)
}
func (m *mockClient) DeleteLoginPassword(token, login string) error {
	return m.deleteLoginPwdFn(token, login)
}
func (m *mockClient) DeleteTextSecret(token, title string) error {
	return m.deleteTextFn(token, title)
}
func (m *mockClient) DeleteBinarySecret(token, filename string) error {
	return m.deleteBinaryFn(token, filename)
}
func (m *mockClient) DeleteCardSecret(token, cardholder string) error {
	return m.deleteCardFn(token, cardholder)
}

// --- mock cache ---

type mockCache struct {
	secrets  *response.AllSecrets
	wrongKey bool
}

func (m *mockCache) Get() *response.AllSecrets { return m.secrets }
func (m *mockCache) Set(s *response.AllSecrets) error {
	m.secrets = s
	return nil
}
func (m *mockCache) Reset()         { m.secrets = nil }
func (m *mockCache) Load() error    { return nil }
func (m *mockCache) Close() error   { return nil }
func (m *mockCache) WrongKey() bool { return m.wrongKey }

// --- Tests ---

func TestGetAllSecrets_ServerAvailable(t *testing.T) {
	expected := &response.AllSecrets{
		LoginPassword: []response.LoginPassword{{Login: "admin"}},
	}
	client := &mockClient{
		getAllSecretsFn: func(_ string) (*response.AllSecrets, error) { return expected, nil },
	}
	cache := &mockCache{}
	uc := usecase.New(client, cache)
	uc.SetToken("tok")

	result, err := uc.GetAllSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LoginPassword) != 1 {
		t.Fatal("expected 1 login-password")
	}
	// Cache should be populated.
	if cache.secrets == nil {
		t.Fatal("expected cache to be populated")
	}
}

func TestGetAllSecrets_OfflineMode(t *testing.T) {
	cached := &response.AllSecrets{
		TextSecret: []response.TextSecret{{Title: "cached-note"}},
	}
	client := &mockClient{
		getAllSecretsFn: func(_ string) (*response.AllSecrets, error) { return nil, errors.New("network error") },
	}
	cache := &mockCache{secrets: cached}
	uc := usecase.New(client, cache)

	result, err := uc.GetAllSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.TextSecret) != 1 || result.TextSecret[0].Title != "cached-note" {
		t.Fatal("expected cached data")
	}
}

func TestGetAllSecrets_OfflineNoCache(t *testing.T) {
	client := &mockClient{
		getAllSecretsFn: func(_ string) (*response.AllSecrets, error) { return nil, errors.New("network error") },
	}
	cache := &mockCache{}
	uc := usecase.New(client, cache)

	_, err := uc.GetAllSecrets()
	if err == nil {
		t.Fatal("expected error when offline with no cache")
	}
}

func TestGetLoginPasswords_OfflineFallback(t *testing.T) {
	cached := &response.AllSecrets{
		LoginPassword: []response.LoginPassword{{Login: "cached"}},
	}
	client := &mockClient{
		getLoginPasswordsFn: func(_ string) ([]response.LoginPassword, error) { return nil, errors.New("offline") },
	}
	cache := &mockCache{secrets: cached}
	uc := usecase.New(client, cache)

	result, err := uc.GetLoginPasswords()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Login != "cached" {
		t.Fatal("expected cached data")
	}
}

func TestAddLoginPassword_Success(t *testing.T) {
	client := &mockClient{
		postLoginPasswordFn: func(_ string, _ request.LoginPassword) error { return nil },
	}
	cache := &mockCache{secrets: &response.AllSecrets{}}
	uc := usecase.New(client, cache)

	err := uc.AddLoginPassword(request.LoginPassword{Login: "x"})
	if err != nil {
		t.Fatal(err)
	}
	// Cache should be reset after write.
	if cache.secrets != nil {
		t.Fatal("expected cache to be reset")
	}
}

func TestAddLoginPassword_ServerError(t *testing.T) {
	client := &mockClient{
		postLoginPasswordFn: func(_ string, _ request.LoginPassword) error { return errors.New("500") },
	}
	cache := &mockCache{}
	uc := usecase.New(client, cache)

	err := uc.AddLoginPassword(request.LoginPassword{Login: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteLoginPassword_Success(t *testing.T) {
	client := &mockClient{
		deleteLoginPwdFn: func(_, _ string) error { return nil },
	}
	cache := &mockCache{secrets: &response.AllSecrets{}}
	uc := usecase.New(client, cache)

	err := uc.DeleteLoginPassword("admin")
	if err != nil {
		t.Fatal(err)
	}
	if cache.secrets != nil {
		t.Fatal("expected cache to be reset")
	}
}

func TestRegister_Success(t *testing.T) {
	client := &mockClient{
		registerFn: func(_, _ string) (string, error) { return "token123", nil },
	}
	cache := &mockCache{}
	uc := usecase.New(client, cache)

	token, err := uc.Register("user", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if token != "token123" {
		t.Fatalf("expected token123, got %s", token)
	}
}

func TestLogin_Success(t *testing.T) {
	client := &mockClient{
		loginFn: func(_, _ string) (string, error) { return "jwt_token", nil },
	}
	cache := &mockCache{}
	uc := usecase.New(client, cache)

	token, err := uc.Login("user", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if token != "jwt_token" {
		t.Fatalf("expected jwt_token, got %s", token)
	}
}

func TestResetCache(t *testing.T) {
	cache := &mockCache{secrets: &response.AllSecrets{}}
	uc := usecase.New(nil, cache)

	uc.ResetCache()
	if cache.secrets != nil {
		t.Fatal("expected cache to be cleared")
	}
}
