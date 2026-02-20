package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"uem-agent/internal/inventory"
	"uem-agent/internal/transport"
)

func main() {
	authURL := getenv("UEM_AUTH_URL", "http://api-gateway:8080/auth/v1/auth/login")
	apiURL := getenv("UEM_DEVICE_URL", "http://api-gateway:8080/devices/v1/devices")
	wsURL := getenv("UEM_WS_URL", "ws://api-gateway:8080/tasks/ws/agent")
	email := getenv("UEM_AGENT_EMAIL", "agent@acme.local")
	password := getenv("UEM_AGENT_PASSWORD", "agent-pass")
	secret := getenv("UEM_HMAC_SECRET", "")

	snapshot := inventory.Collect()
	body, _ := json.Marshal(snapshot)

	for i := 1; i <= 5; i++ {
		token, err := transport.Authenticate(authURL, email, password)
		if err != nil {
			log.Printf("auth attempt %d failed: %v", i, err)
			time.Sleep(3 * time.Second)
			continue
		}

		if err := transport.SendHTTPS(apiURL, body, secret, token); err != nil {
			log.Printf("https attempt %d failed, fallback websocket: %v", i, err)
			if wsErr := transport.SendWebSocket(wsURL, body); wsErr != nil {
				log.Printf("websocket attempt %d failed: %v", i, wsErr)
				time.Sleep(3 * time.Second)
				continue
			}
		}

		log.Println("inventory sent")
		return
	}

	log.Fatal("all retry attempts failed")
}

func getenv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
