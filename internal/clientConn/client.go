// Package clientconn implements the HTTP client that communicates with
// the GophKeeper server using Fiber's built-in HTTP agent.
//
// All request bodies are encrypted with AES-256-GCM before being sent,
// and all response bodies are decrypted upon receipt. The same symmetric
// key is used on both sides (client and server CryptoMiddleware).
//
// The package exposes a single [Client] struct that satisfies the
// usecase.HTTPClient interface.
package clientconn

import (
	"encoding/json"
	"fmt"

	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/crypto"
	"github.com/gofiber/fiber/v2"
)

// Client is a Fiber-based HTTP client for the GophKeeper REST API.
type Client struct {
	baseURL   string // e.g. "http://127.0.0.1:8080/v1/api/user"
	cryptoKey []byte // 32-byte AES key derived from the passphrase
}

// New creates a Client pointing at the given host:port.
// The cryptoKey string is hashed with SHA-256 to produce a 32-byte AES key.
func New(host, port, cryptoKey string) *Client {
	return &Client{
		baseURL:   fmt.Sprintf("http://%s:%s/v1/api/user", host, port),
		cryptoKey: crypto.DeriveKey(cryptoKey),
	}
}

// Register creates a new user account on the server.
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

// Login authenticates and returns a JWT token string.
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

// GetAllSecrets fetches every secret owned by the authenticated user.
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

// PostLoginPassword creates a login-password secret on the server.
func (c *Client) PostLoginPassword(token string, lp request.LoginPassword) error {
	return c.writeOp(c.baseURL+"/secret/post-login-password", token, lp)
}

// PostTextSecret creates a text secret on the server.
func (c *Client) PostTextSecret(token string, ts request.TextSecret) error {
	return c.writeOp(c.baseURL+"/secret/post-text-secret", token, ts)
}

// PostBinarySecret creates a binary secret on the server.
func (c *Client) PostBinarySecret(token string, bs request.BinarySecret) error {
	return c.writeOp(c.baseURL+"/secret/post-binary-secret", token, bs)
}

// PostCardSecret creates a card secret on the server.
func (c *Client) PostCardSecret(token string, cs request.CardSecret) error {
	return c.writeOp(c.baseURL+"/secret/post-card-secret", token, cs)
}

// DeleteLoginPassword removes a login-password by its login identifier.
func (c *Client) DeleteLoginPassword(token, login string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-login-password", token, request.DeleteLoginPassword{Login: login})
}

// DeleteTextSecret removes a text secret by its title.
func (c *Client) DeleteTextSecret(token, title string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-text-secret", token, request.DeleteTextSecret{Title: title})
}

// DeleteBinarySecret removes a binary secret by filename.
func (c *Client) DeleteBinarySecret(token, filename string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-binary-secret", token, request.DeleteBinarySecret{Filename: filename})
}

// DeleteCardSecret removes a card secret by cardholder name.
func (c *Client) DeleteCardSecret(token, cardholder string) error {
	return c.deleteOp(c.baseURL+"/secret/delete-card-secret", token, request.DeleteCardSecret{Cardholder: cardholder})
}

// --- internal helpers ---

// serverError tries to extract an {"error":"..."} message from the
// decrypted response body. If parsing fails, returns a generic error
// with the HTTP status code.
func (c *Client) serverError(code int, body []byte) error {
	var resp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &resp) == nil && resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return fmt.Errorf("server error %d", code)
}

// encryptBody JSON-marshals v and encrypts the result with AES-256-GCM.
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

// decryptBody decrypts the raw server response. If decryption fails
// (e.g. the response is not encrypted), the raw bytes are returned as-is
// to allow the caller to handle plaintext error messages.
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

// doPost sends an encrypted POST request.
// If token is non-empty, an Authorization: Bearer header is added.
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

// doGet sends a GET request (no body) and decrypts the response.
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

// writeOp is a shorthand for POST endpoints that return only a status
// message (no data to parse beyond error checking).
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

// deleteOp is a shorthand for DELETE endpoints that carry a JSON body
// identifying the resource to remove.
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
