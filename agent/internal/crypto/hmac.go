package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func SignPayload(body []byte, nonce, timestamp, secret string) string {
	message := fmt.Sprintf("%s|%s|%s", string(body), nonce, timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
