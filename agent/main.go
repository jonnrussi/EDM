package main

import (
	"encoding/json"
	"log"
	"os"

	"uem-agent/internal/auth"
	"uem-agent/internal/inventory"
	"uem-agent/internal/transport"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("agent config error: %v", err)
	}

	authClient := &auth.Client{URL: cfg.AuthURL, Email: cfg.AgentEmail, Password: cfg.AgentPassword}
	token, err := authClient.GetToken()
	if err != nil {
		log.Fatalf("jwt login failed: %v", err)
	}

	snapshot := inventory.Collect()
	body, _ := json.Marshal(snapshot)

	status, sendErr := transport.SendHTTPS(cfg.DeviceURL, token, body, cfg.HMACSecret)
	if sendErr != nil && status == 401 {
		log.Printf("jwt expired/invalid, retrying login: %v", sendErr)
		if err := authClient.ForceRelogin(); err != nil {
			log.Fatalf("jwt refresh failed: %v", err)
		}
		token, _ = authClient.GetToken()
		status, sendErr = transport.SendHTTPS(cfg.DeviceURL, token, body, cfg.HMACSecret)
	}

	if sendErr != nil {
		if status == 403 {
			log.Printf("hmac or schema validation failure: %v", sendErr)
		}
		log.Printf("https failed, fallback websocket: %v", sendErr)
		if wsErr := transport.SendWebSocket(cfg.WSURL, body); wsErr != nil {
			log.Fatalf("all transports failed: %v", wsErr)
		}
	}
	log.Println("inventory sent")
}

type config struct {
	AuthURL       string
	DeviceURL     string
	WSURL         string
	HMACSecret    string
	AgentEmail    string
	AgentPassword string
}

func loadConfig() (*config, error) {
	cfg := &config{
		AuthURL:       getenv("UEM_AUTH_URL", "http://localhost:8001/v1/auth/login"),
		DeviceURL:     getenv("UEM_DEVICE_URL", "http://localhost:8002/v1/devices"),
		WSURL:         getenv("UEM_WS_URL", "ws://localhost:8003/ws/agent"),
		HMACSecret:    os.Getenv("UEM_HMAC_SECRET"),
		AgentEmail:    os.Getenv("UEM_AGENT_EMAIL"),
		AgentPassword: os.Getenv("UEM_AGENT_PASSWORD"),
	}

	if cfg.HMACSecret == "" || cfg.AgentEmail == "" || cfg.AgentPassword == "" {
		return nil, os.ErrInvalid
	}
	return cfg, nil
}

func getenv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
