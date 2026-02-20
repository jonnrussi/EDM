package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"uem-agent/internal/transport"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type jwtClaims struct {
	Exp int64 `json:"exp"`
}

type TokenManager struct {
	client        *transport.Client
	loginURL      string
	email         string
	password      string
	refreshLeeway time.Duration

	mu    sync.RWMutex
	token string
	exp   time.Time
}

func NewTokenManager(client *transport.Client, loginURL, email, password string, refreshLeeway time.Duration) *TokenManager {
	return &TokenManager{client: client, loginURL: loginURL, email: email, password: password, refreshLeeway: refreshLeeway}
}

func (m *TokenManager) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = m.EnsureValidToken(ctx)
			}
		}
	}()
}

func (m *TokenManager) Token() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.token
}

func (m *TokenManager) EnsureValidToken(ctx context.Context) error {
	m.mu.RLock()
	shouldRefresh := m.token == "" || time.Until(m.exp) <= m.refreshLeeway
	m.mu.RUnlock()
	if !shouldRefresh {
		return nil
	}
	return m.Login(ctx)
}

func (m *TokenManager) Login(ctx context.Context) error {
	if m.email == "" || m.password == "" {
		return errors.New("missing agent credentials")
	}
	var out loginResponse
	err := m.client.DoJSON(ctx, transport.Request{
		Method:  "POST",
		URL:     m.loginURL,
		Body:    loginRequest{Email: m.email, Password: m.password},
		UseAuth: false,
	}, &out)
	if err != nil {
		return err
	}
	if out.AccessToken == "" {
		return errors.New("login returned empty token")
	}
	exp, err := parseJWTExp(out.AccessToken)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.token = out.AccessToken
	m.exp = exp
	m.mu.Unlock()
	return nil
}

func (m *TokenManager) ClearToken() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = ""
	m.exp = time.Time{}
}

func parseJWTExp(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid jwt")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, err
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, err
	}
	if claims.Exp == 0 {
		return time.Time{}, fmt.Errorf("jwt without exp")
	}
	return time.Unix(claims.Exp, 0), nil
}
