package transport

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"net/http"
	"os"
	"time"

	"uem-agent/internal/crypto"
)

type Request struct {
	Method  string
	URL     string
	Body    any
	UseAuth bool
}

type Client struct {
	httpClient     *http.Client
	hmacSecret     string
	maxAttempts    int
	baseDelay      time.Duration
	maxDelay       time.Duration
	tokenProvider  func() string
	onUnauthorized func(context.Context) error
}

func NewClient(certPath, keyPath, caPath, hmacSecret string, mtlsDisabled bool, timeout time.Duration, maxAttempts int, baseDelay, maxDelay time.Duration) (*Client, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if !mtlsDisabled {
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
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &Client{
		httpClient:  &http.Client{Timeout: timeout, Transport: transport},
		hmacSecret:  hmacSecret,
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
		maxDelay:    maxDelay,
	}, nil
}

func (c *Client) SetTokenProvider(fn func() string)                     { c.tokenProvider = fn }
func (c *Client) SetUnauthorizedHandler(fn func(context.Context) error) { c.onUnauthorized = fn }

func (c *Client) DoJSON(ctx context.Context, in Request, out any) error {
	bodyBytes, err := marshalBody(in.Body)
	if err != nil {
		return err
	}

	attempts := max(1, c.maxAttempts)
	for attempt := 1; attempt <= attempts; attempt++ {
		status, payload, err := c.doOnce(ctx, in, bodyBytes)
		if err == nil && status < 300 {
			if out != nil && len(payload) > 0 {
				return json.Unmarshal(payload, out)
			}
			return nil
		}

		if status == http.StatusUnauthorized && c.onUnauthorized != nil {
			if reErr := c.onUnauthorized(ctx); reErr == nil {
				continue
			}
		}

		if attempt == attempts || !retryable(status, err) {
			if err != nil {
				return err
			}
			return fmt.Errorf("http failure: %d, body=%s", status, string(payload))
		}
		sleepWithJitter(backoffDuration(c.baseDelay, c.maxDelay, attempt), attempt)
	}
	return errors.New("unreachable retry loop")
}

func (c *Client) doOnce(ctx context.Context, in Request, bodyBytes []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, in.Method, in.URL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, nil, err
	}
	nonce := randomHex(12)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", randomHex(8))
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", crypto.SignPayload(bodyBytes, nonce, timestamp, c.hmacSecret))
	if in.UseAuth && c.tokenProvider != nil {
		if t := c.tokenProvider(); t != "" {
			req.Header.Set("Authorization", "Bearer "+t)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, payload, nil
}

func marshalBody(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return json.Marshal(body)
}

func retryable(status int, err error) bool {
	if err != nil {
		return true
	}
	return status == 429 || status >= 500
}

func backoffDuration(base, maxDelay time.Duration, attempt int) time.Duration {
	if attempt <= 1 {
		return base
	}
	factor := math.Pow(2, float64(attempt-1))
	candidate := time.Duration(float64(base) * factor)
	if candidate > maxDelay {
		return maxDelay
	}
	return candidate
}

func sleepWithJitter(d time.Duration, attempt int) {
	s := mrand.New(mrand.NewSource(time.Now().UnixNano() + int64(attempt)))
	jitter := time.Duration(s.Int63n(int64(max(100*time.Millisecond, d/2))))
	time.Sleep(d + jitter)
}

func randomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func max[T ~int64 | ~int](a, b T) T {
	if a > b {
		return a
	}
	return b
}
