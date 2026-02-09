package clientconn

import (
	"encoding/json"
	"fmt"

	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/gofiber/fiber/v2"
)

type Client struct {
	baseURL string
}

func New(host, port string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%s/v1/api/user", host, port),
	}
}

func (c *Client) Register(login, password string) error {
	a := fiber.Post(c.baseURL + "/register")
	a.JSON(request.UserInput{Login: login, Password: password})

	code, body, errs := a.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}
	if code == fiber.StatusConflict {
		return fmt.Errorf("user already exists")
	}
	if code != fiber.StatusOK {
		return fmt.Errorf("register: %d %s", code, string(body))
	}
	return nil
}

func (c *Client) Login(login, password string) (string, error) {
	a := fiber.Post(c.baseURL + "/login")
	a.JSON(request.UserInput{Login: login, Password: password})

	code, body, errs := a.Bytes()
	if len(errs) > 0 {
		return "", errs[0]
	}
	if code == fiber.StatusUnauthorized {
		return "", fmt.Errorf("invalid credentials")
	}
	if code != fiber.StatusOK {
		return "", fmt.Errorf("login: %d %s", code, string(body))
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
	a := fiber.Get(c.baseURL + "/secret/get-all-secrets")
	a.Set("Authorization", "Bearer "+token)

	code, body, errs := a.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	if code == fiber.StatusUnauthorized {
		return nil, fmt.Errorf("session expired, please re-login")
	}
	if code != fiber.StatusOK {
		return nil, fmt.Errorf("get secrets: %d %s", code, string(body))
	}

	var secrets response.AllSecrets
	if err := json.Unmarshal(body, &secrets); err != nil {
		return nil, fmt.Errorf("parse secrets: %w", err)
	}
	return &secrets, nil
}

func (c *Client) PostLoginPassword(token string, lp request.LoginPassword) error {
	a := fiber.Post(c.baseURL + "/secret/post-login-password")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(lp)
	return c.checkWrite(a)
}

func (c *Client) PostTextSecret(token string, ts request.TextSecret) error {
	a := fiber.Post(c.baseURL + "/secret/post-text-secret")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(ts)
	return c.checkWrite(a)
}

func (c *Client) PostBinarySecret(token string, bs request.BinarySecret) error {
	a := fiber.Post(c.baseURL + "/secret/post-binary-secret")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(bs)
	return c.checkWrite(a)
}

func (c *Client) PostCardSecret(token string, cs request.CardSecret) error {
	a := fiber.Post(c.baseURL + "/secret/post-card-secret")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(cs)
	return c.checkWrite(a)
}

func (c *Client) DeleteLoginPassword(token, login string) error {
	a := fiber.Delete(c.baseURL + "/secret/delete-login-password")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(request.DeleteLoginPassword{Login: login})
	return c.checkWrite(a)
}

func (c *Client) DeleteTextSecret(token, title string) error {
	a := fiber.Delete(c.baseURL + "/secret/delete-text-secret")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(request.DeleteTextSecret{Title: title})
	return c.checkWrite(a)
}

func (c *Client) DeleteBinarySecret(token, filename string) error {
	a := fiber.Delete(c.baseURL + "/secret/delete-binary-secret")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(request.DeleteBinarySecret{Filename: filename})
	return c.checkWrite(a)
}

func (c *Client) DeleteCardSecret(token, cardholder string) error {
	a := fiber.Delete(c.baseURL + "/secret/delete-card-secret")
	a.Set("Authorization", "Bearer "+token)
	a.JSON(request.DeleteCardSecret{Cardholder: cardholder})
	return c.checkWrite(a)
}

func (c *Client) checkWrite(a *fiber.Agent) error {
	code, body, errs := a.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}
	if code == fiber.StatusUnauthorized {
		return fmt.Errorf("session expired, please re-login")
	}
	if code >= 400 {
		return fmt.Errorf("server error %d: %s", code, string(body))
	}
	return nil
}
