package transport

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"uem-agent/internal/crypto"
)

type Client struct {
	httpClient *http.Client
	hmacSecret string
}

func NewClient(certPath, keyPath, caPath, hmacSecret string) (*Client, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load device certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if caPath != "" {
		caBytes, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("invalid CA bundle")
		}
		tlsConfig.RootCAs = pool
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second, Transport: transport},
		hmacSecret: hmacSecret,
	}, nil
}

func (c *Client) DoJSON(method, url string, reqBody any, out any) error {
	var bodyBytes []byte
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyBytes = data
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", crypto.Sign(bodyBytes, c.hmacSecret))
	req.Header.Set("X-Nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("X-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http failure: %d, body=%s", resp.StatusCode, string(payload))
	}
	if out != nil && len(payload) > 0 {
		if err := json.Unmarshal(payload, out); err != nil {
			return err
		}
	}
	return nil
}
