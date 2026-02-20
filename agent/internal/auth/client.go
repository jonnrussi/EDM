package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	URL      string
	Email    string
	Password string
	mu       sync.Mutex
	token    string
	expires  time.Time
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

func (c *Client) GetToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expires.Add(-15*time.Second)) {
		return c.token, nil
	}

	if err := c.login(); err != nil {
		return "", err
	}
	return c.token, nil
}

func (c *Client) ForceRelogin() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.login()
}

func (c *Client) login() error {
	payload, _ := json.Marshal(loginRequest{Email: c.Email, Password: c.Password})
	resp, err := http.Post(c.URL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("auth login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("auth login failed with status %d", resp.StatusCode)
	}

	var body loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("auth response decode failed: %w", err)
	}

	exp, err := parseJWTExp(body.AccessToken)
	if err != nil {
		return err
	}
	c.token = body.AccessToken
	c.expires = exp
	return nil
}

func parseJWTExp(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("jwt malformed")
	}

	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("jwt payload decode failed: %w", err)
	}

	var payload struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return time.Time{}, fmt.Errorf("jwt payload parse failed: %w", err)
	}
	if payload.Exp == 0 {
		return time.Time{}, fmt.Errorf("jwt exp missing")
	}

	return time.Unix(payload.Exp, 0), nil
}
