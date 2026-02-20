package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"uem-agent/internal/crypto"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func Authenticate(authURL, email, password string) (string, error) {
	body, _ := json.Marshal(LoginRequest{Email: email, Password: password})
	req, _ := http.NewRequest(http.MethodPost, authURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth failed: %d %s", resp.StatusCode, string(payload))
	}
	var loginResponse LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return "", err
	}
	if loginResponse.AccessToken == "" {
		return "", fmt.Errorf("auth failed: empty token")
	}
	return loginResponse.AccessToken, nil
}

func SendHTTPS(url string, body []byte, secret string, token string) error {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Signature", crypto.Sign(body, secret))
	req.Header.Set("X-Nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("X-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http failure: %d", resp.StatusCode)
	}
	return nil
}

func SendWebSocket(wsURL string, body []byte) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.WriteMessage(websocket.TextMessage, body)
}
