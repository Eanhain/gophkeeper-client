package usecase

import (
	"errors"
	"testing"

	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
)

// --- mocks ---

type mockClient struct {
	registerFn          func(login, password string) error
	loginFn             func(login, password string) (string, error)
	getAllSecretsFn      func(token string) (*response.AllSecrets, error)
	postLoginPasswordFn func(token string, lp request.LoginPassword) error
	postTextSecretFn    func(token string, ts request.TextSecret) error
	postBinarySecretFn  func(token string, bs request.BinarySecret) error
	postCardSecretFn    func(token string, cs request.CardSecret) error
	deleteLoginPassFn   func(token, login string) error
	deleteTextFn        func(token, title string) error
	deleteBinaryFn      func(token, filename string) error
	deleteCardFn        func(token, cardholder string) error
}

func (m *mockClient) Register(l, p string) error              { return m.registerFn(l, p) }
func (m *mockClient) Login(l, p string) (string, error)       { return m.loginFn(l, p) }
func (m *mockClient) GetAllSecrets(t string) (*response.AllSecrets, error) {
	return m.getAllSecretsFn(t)
}
func (m *mockClient) PostLoginPassword(t string, lp request.LoginPassword) error {
	return m.postLoginPasswordFn(t, lp)
}
func (m *mockClient) PostTextSecret(t string, ts request.TextSecret) error {
	return m.postTextSecretFn(t, ts)
}
func (m *mockClient) PostBinarySecret(t string, bs request.BinarySecret) error {
	return m.postBinarySecretFn(t, bs)
}
func (m *mockClient) PostCardSecret(t string, cs request.CardSecret) error {
	return m.postCardSecretFn(t, cs)
}
func (m *mockClient) DeleteLoginPassword(t, l string) error   { return m.deleteLoginPassFn(t, l) }
func (m *mockClient) DeleteTextSecret(t, title string) error   { return m.deleteTextFn(t, title) }
func (m *mockClient) DeleteBinarySecret(t, f string) error     { return m.deleteBinaryFn(t, f) }
func (m *mockClient) DeleteCardSecret(t, ch string) error      { return m.deleteCardFn(t, ch) }

type mockCache struct {
	data *response.AllSecrets
}

func (m *mockCache) Get() *response.AllSecrets          { return m.data }
func (m *mockCache) Set(s *response.AllSecrets) error   { m.data = s; return nil }
func (m *mockCache) Reset()                             { m.data = nil }

// --- tests ---

func TestLogin(t *testing.T) {
	mc := &mockClient{loginFn: func(l, p string) (string, error) { return "jwt-token", nil }}
	uc := New(mc, &mockCache{})

	tok, err := uc.Login("user", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "jwt-token" {
		t.Fatalf("expected jwt-token, got %s", tok)
	}
}

func TestLoginError(t *testing.T) {
	mc := &mockClient{loginFn: func(l, p string) (string, error) { return "", errors.New("fail") }}
	uc := New(mc, &mockCache{})

	_, err := uc.Login("user", "pass")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegister(t *testing.T) {
	mc := &mockClient{
		registerFn: func(l, p string) error { return nil },
		loginFn:    func(l, p string) (string, error) { return "tok", nil },
	}
	uc := New(mc, &mockCache{})

	tok, err := uc.Register("u", "p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "tok" {
		t.Fatalf("expected tok, got %s", tok)
	}
}

func TestRegisterFail(t *testing.T) {
	mc := &mockClient{registerFn: func(l, p string) error { return errors.New("conflict") }}
	uc := New(mc, &mockCache{})

	_, err := uc.Register("u", "p")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAllSecretsCached(t *testing.T) {
	secrets := &response.AllSecrets{LoginPassword: []response.LoginPassword{{Login: "a"}}}
	cache := &mockCache{data: secrets}
	uc := New(&mockClient{}, cache)

	got, err := uc.GetAllSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LoginPassword[0].Login != "a" {
		t.Fatal("expected cached data")
	}
}

func TestGetAllSecretsFromServer(t *testing.T) {
	secrets := &response.AllSecrets{TextSecret: []response.TextSecret{{Title: "note"}}}
	mc := &mockClient{getAllSecretsFn: func(t string) (*response.AllSecrets, error) { return secrets, nil }}
	cache := &mockCache{}
	uc := New(mc, cache)
	uc.SetToken("tok")

	got, err := uc.GetAllSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TextSecret[0].Title != "note" {
		t.Fatal("expected server data")
	}
	if cache.data == nil {
		t.Fatal("expected data to be cached")
	}
}

func TestGetAllSecretsServerError(t *testing.T) {
	mc := &mockClient{getAllSecretsFn: func(t string) (*response.AllSecrets, error) {
		return nil, errors.New("fail")
	}}
	uc := New(mc, &mockCache{})

	_, err := uc.GetAllSecrets()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddLoginPassword(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{postLoginPasswordFn: func(t string, lp request.LoginPassword) error { return nil }}
	uc := New(mc, cache)

	if err := uc.AddLoginPassword(request.LoginPassword{Login: "a"}); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset after write")
	}
}

func TestAddLoginPasswordError(t *testing.T) {
	mc := &mockClient{postLoginPasswordFn: func(t string, lp request.LoginPassword) error {
		return errors.New("fail")
	}}
	uc := New(mc, &mockCache{})

	if err := uc.AddLoginPassword(request.LoginPassword{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddTextSecret(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{postTextSecretFn: func(t string, ts request.TextSecret) error { return nil }}
	uc := New(mc, cache)

	if err := uc.AddTextSecret(request.TextSecret{Title: "t"}); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestAddTextSecretError(t *testing.T) {
	mc := &mockClient{postTextSecretFn: func(t string, ts request.TextSecret) error {
		return errors.New("fail")
	}}
	uc := New(mc, &mockCache{})
	if err := uc.AddTextSecret(request.TextSecret{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddBinarySecret(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{postBinarySecretFn: func(t string, bs request.BinarySecret) error { return nil }}
	uc := New(mc, cache)

	if err := uc.AddBinarySecret(request.BinarySecret{Filename: "f"}); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestAddBinarySecretError(t *testing.T) {
	mc := &mockClient{postBinarySecretFn: func(t string, bs request.BinarySecret) error {
		return errors.New("fail")
	}}
	uc := New(mc, &mockCache{})
	if err := uc.AddBinarySecret(request.BinarySecret{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddCardSecret(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{postCardSecretFn: func(t string, cs request.CardSecret) error { return nil }}
	uc := New(mc, cache)

	if err := uc.AddCardSecret(request.CardSecret{Cardholder: "c"}); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestAddCardSecretError(t *testing.T) {
	mc := &mockClient{postCardSecretFn: func(t string, cs request.CardSecret) error {
		return errors.New("fail")
	}}
	uc := New(mc, &mockCache{})
	if err := uc.AddCardSecret(request.CardSecret{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteLoginPassword(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{deleteLoginPassFn: func(t, l string) error { return nil }}
	uc := New(mc, cache)

	if err := uc.DeleteLoginPassword("admin"); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestDeleteLoginPasswordError(t *testing.T) {
	mc := &mockClient{deleteLoginPassFn: func(t, l string) error { return errors.New("fail") }}
	uc := New(mc, &mockCache{})
	if err := uc.DeleteLoginPassword("x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteTextSecret(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{deleteTextFn: func(t, title string) error { return nil }}
	uc := New(mc, cache)

	if err := uc.DeleteTextSecret("note"); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestDeleteTextSecretError(t *testing.T) {
	mc := &mockClient{deleteTextFn: func(t, title string) error { return errors.New("fail") }}
	uc := New(mc, &mockCache{})
	if err := uc.DeleteTextSecret("x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteBinarySecret(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{deleteBinaryFn: func(t, f string) error { return nil }}
	uc := New(mc, cache)

	if err := uc.DeleteBinarySecret("file.bin"); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestDeleteBinarySecretError(t *testing.T) {
	mc := &mockClient{deleteBinaryFn: func(t, f string) error { return errors.New("fail") }}
	uc := New(mc, &mockCache{})
	if err := uc.DeleteBinarySecret("x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteCardSecret(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	mc := &mockClient{deleteCardFn: func(t, ch string) error { return nil }}
	uc := New(mc, cache)

	if err := uc.DeleteCardSecret("John"); err != nil {
		t.Fatal(err)
	}
	if cache.data != nil {
		t.Fatal("cache should be reset")
	}
}

func TestDeleteCardSecretError(t *testing.T) {
	mc := &mockClient{deleteCardFn: func(t, ch string) error { return errors.New("fail") }}
	uc := New(mc, &mockCache{})
	if err := uc.DeleteCardSecret("x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestResetCache(t *testing.T) {
	cache := &mockCache{data: &response.AllSecrets{}}
	uc := New(&mockClient{}, cache)

	uc.ResetCache()
	if cache.data != nil {
		t.Fatal("cache should be nil after reset")
	}
}

func TestSetToken(t *testing.T) {
	uc := New(&mockClient{}, &mockCache{})
	uc.SetToken("abc")
	if uc.token != "abc" {
		t.Fatal("token not set")
	}
}
