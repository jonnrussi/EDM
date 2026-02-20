package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"uem-agent/internal/crypto"
)

func SendHTTPS(url, token string, body []byte, secret string) (int, error) {
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := crypto.Sign(body, nonce, timestamp, secret)

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Timestamp", timestamp)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("http failure: %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func SendWebSocket(wsURL string, body []byte) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.WriteMessage(websocket.TextMessage, body)
}
