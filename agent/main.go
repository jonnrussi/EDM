package main

import (
	"encoding/json"
	"log"
	"os"

	"uem-agent/internal/inventory"
	"uem-agent/internal/transport"
)

func main() {
	apiURL := getenv("UEM_DEVICE_URL", "https://localhost:8002/v1/devices")
	wsURL := getenv("UEM_WS_URL", "ws://localhost:8003/ws/agent")
	secret := getenv("UEM_HMAC_SECRET", "replace")

	snapshot := inventory.Collect()
	body, _ := json.Marshal(snapshot)
	if err := transport.SendHTTPS(apiURL, body, secret); err != nil {
		log.Printf("https failed, fallback websocket: %v", err)
		if wsErr := transport.SendWebSocket(wsURL, body); wsErr != nil {
			log.Fatalf("all transports failed: %v", wsErr)
		}
	}
	log.Println("inventory sent")
}

func getenv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
