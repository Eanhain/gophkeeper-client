package clientconn

import (
	"encoding/json"
	"fmt"

	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/crypto"
	"github.com/gofiber/fiber/v2"
)

type Client struct {
	baseURL   string
	cryptoKey []byte
}

func New(host, port, cryptoKey string) *Client {
	return &Client{
		baseURL:   fmt.Sprintf("http://%s:%s/v1/api/user", host, port),
		cryptoKey: crypto.DeriveKey(cryptoKey),
	}
}

func (c *Client) Register(login, password string) error {
	code, body, err := c.doPost(c.baseURL+"/register", "", request.UserInput{Login: login, Password: password})
	if err != nil {
		return err
	}
	if code != fiber.StatusOK {
		return c.serverError(code, body)
	}
	return nil
}

func (c *Client) Login(login, password string) (string, error) {
	code, body, err := c.doPost(c.baseURL+"/login", "", request.UserInput{Login: login, Password: password})
	if err != nil {
		return "", err
	}
	if code != fiber.StatusOK {
		return "", c.serverError(code, body)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	return result.Token, nil
}

func (c *Client) GetAllSecrets(token string) (*response.AllSecrets, error) {
	code, body, err := c.doGet(c.baseURL+"/secret/get-all-secrets", token)
	if err != nil {
		return nil, err
	}
	if code != fiber.StatusOK {
		return nil, c.serverError(code, body)
	}

	var secrets response.AllSecrets
	if err := json.Unmarshal(body, &secrets); err != nil {
		return nil, fmt.Errorf("parse secrets: %w", err)
	}
	return &secrets, nil
}

func (c *Client) PostLoginPassword(token string, lp request.LoginPassword) error {
	return c.writeOp(c.baseURL+"/secret/post-login-password", token, lp)
}

func (c *Client) PostTextSecret(token string, ts request.TextSecret) error {
	return c.writeOp(c.baseURL+"/secret/post-text-secret", token, ts)
}

func (c *Client) PostBinarySecret(token string, bs request.BinarySecret) error {
	return c.writeOp(c.baseURL+"/secret/post-binary-secret", token, bs)
}

func (c *Client) PostCardSecret(token string, cs request.CardSecret) error {
	return c.writeOp(c.baseURL+"/secret/post-card-secret", token, cs)
}

func (c *Client) DeleteLoginPassword(token, login string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-login-password", token, request.DeleteLoginPassword{Login: login})
}

func (c *Client) DeleteTextSecret(token, title string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-text-secret", token, request.DeleteTextSecret{Title: title})
}

func (c *Client) DeleteBinarySecret(token, filename string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-binary-secret", token, request.DeleteBinarySecret{Filename: filename})
}

func (c *Client) DeleteCardSecret(token, cardholder string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-card-secret", token, request.DeleteCardSecret{Cardholder: cardholder})
}

func (c *Client) serverError(code int, body []byte) error {
	var resp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &resp) == nil && resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return fmt.Errorf("server error %d", code)
}

func (c *Client) encryptBody(v any) ([]byte, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	enc, err := crypto.Encrypt(c.cryptoKey, j)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	return enc, nil
}

func (c *Client) decryptBody(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	plain, err := crypto.Decrypt(c.cryptoKey, raw)
	if err != nil {
		return raw, nil
	}
	return plain, nil
}

func (c *Client) doPost(url, token string, body any) (int, []byte, error) {
	enc, err := c.encryptBody(body)
	if err != nil {
		return 0, nil, err
	}

	a := fiber.Post(url)
	if token != "" {
		a.Set("Authorization", "Bearer "+token)
	}
	a.ContentType(fiber.MIMEOctetStream)
	a.Body(enc)

	code, raw, errs := a.Bytes()
	if len(errs) > 0 {
		return 0, nil, errs[0]
	}
	plain, _ := c.decryptBody(raw)
	return code, plain, nil
}

func (c *Client) doGet(url, token string) (int, []byte, error) {
	a := fiber.Get(url)
	a.Set("Authorization", "Bearer "+token)

	code, raw, errs := a.Bytes()
	if len(errs) > 0 {
		return 0, nil, errs[0]
	}
	plain, _ := c.decryptBody(raw)
	return code, plain, nil
}

func (c *Client) writeOp(url, token string, body any) error {
	code, respBody, err := c.doPost(url, token, body)
	if err != nil {
		return err
	}
	if code >= 400 {
		return c.serverError(code, respBody)
	}
	return nil
}

func (c *Client) deleteOp(url, token string, body any) error {
	enc, err := c.encryptBody(body)
	if err != nil {
		return err
	}

	a := fiber.Delete(url)
	a.Set("Authorization", "Bearer "+token)
	a.ContentType(fiber.MIMEOctetStream)
	a.Body(enc)

	code, raw, errs := a.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}
	if code >= 400 {
		plain, _ := c.decryptBody(raw)
		return c.serverError(code, plain)
	}
	return nil
}
