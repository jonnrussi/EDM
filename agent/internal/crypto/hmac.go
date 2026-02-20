package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Sign(body []byte, nonce, timestamp, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	mac.Write([]byte(nonce))
	mac.Write([]byte(timestamp))
	return hex.EncodeToString(mac.Sum(nil))
}
