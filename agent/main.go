package main

import (
	"log"
	"os"
	"strconv"

	"uem-agent/internal/service"
	"uem-agent/internal/transport"
)

func main() {
	cfg := service.Config{
		DeviceID:        getenv("UEM_DEVICE_ID", "device-local"),
		DeviceURL:       getenv("UEM_DEVICE_URL", "https://localhost:8002/v1/devices"),
		TaskServiceURL:  getenv("UEM_TASK_URL", "https://localhost:8003"),
		PollIntervalSec: getenvInt("UEM_POLL_INTERVAL_SEC", 15),
		AllowedCommands: getenv("UEM_ALLOWED_COMMANDS", "echo,hostname,uname"),
	}

	client, err := transport.NewClient(
		os.Getenv("UEM_CLIENT_CERT_PATH"),
		os.Getenv("UEM_CLIENT_KEY_PATH"),
		os.Getenv("UEM_CA_CERT_PATH"),
		getenv("UEM_HMAC_SECRET", "replace"),
	)
	if err != nil {
		log.Fatalf("unable to initialize secure transport: %v", err)
	}

	agent := service.New(cfg, client)
	agent.Run()
}

func getenv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func getenvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return parsed
}
