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

func SendHTTPS(url string, body []byte, secret string) error {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
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
