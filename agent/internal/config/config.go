package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AuthURL          string
	EnrollmentURL    string
	DeviceURL        string
	TaskServiceURL   string
	AgentEmail       string
	AgentPassword    string
	EnrollmentToken  string
	DeviceIDFile     string
	HMACSecret       string
	MTLSDisabled     bool
	ClientCertPath   string
	ClientKeyPath    string
	CACertPath       string
	AllowedCommands  map[string]struct{}
	PollInterval     time.Duration
	ExecutionTimeout time.Duration
	RequestTimeout   time.Duration
	RetryMaxAttempts int
	RetryBaseDelay   time.Duration
	RetryMaxDelay    time.Duration
	RefreshLeeway    time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AuthURL:          getenv("UEM_AUTH_URL", "https://localhost:8001/v1/auth/login"),
		EnrollmentURL:    getenv("UEM_ENROLLMENT_URL", "https://localhost:8001/v1/agents/enroll"),
		DeviceURL:        getenv("UEM_DEVICE_URL", "https://localhost:8002/v1/devices"),
		TaskServiceURL:   getenv("UEM_TASK_URL", "https://localhost:8003"),
		AgentEmail:       os.Getenv("UEM_AGENT_EMAIL"),
		AgentPassword:    os.Getenv("UEM_AGENT_PASSWORD"),
		EnrollmentToken:  os.Getenv("UEM_ENROLLMENT_TOKEN"),
		DeviceIDFile:     getenv("UEM_DEVICE_ID_FILE", ".uem_device_id"),
		HMACSecret:       getenv("UEM_HMAC_SECRET", "replace"),
		MTLSDisabled:     getenvBool("MTLS_DISABLED", false),
		ClientCertPath:   os.Getenv("UEM_CLIENT_CERT_PATH"),
		ClientKeyPath:    os.Getenv("UEM_CLIENT_KEY_PATH"),
		CACertPath:       os.Getenv("UEM_CA_CERT_PATH"),
		PollInterval:     getenvDuration("UEM_POLL_INTERVAL", 15*time.Second),
		ExecutionTimeout: getenvDuration("UEM_EXECUTION_TIMEOUT", 60*time.Second),
		RequestTimeout:   getenvDuration("UEM_REQUEST_TIMEOUT", 20*time.Second),
		RetryMaxAttempts: getenvInt("UEM_RETRY_MAX_ATTEMPTS", 3),
		RetryBaseDelay:   getenvDuration("UEM_RETRY_BASE_DELAY", 500*time.Millisecond),
		RetryMaxDelay:    getenvDuration("UEM_RETRY_MAX_DELAY", 8*time.Second),
		RefreshLeeway:    getenvDuration("UEM_TOKEN_REFRESH_LEEWAY", 60*time.Second),
	}
	cfg.AllowedCommands = map[string]struct{}{}
	for _, cmd := range strings.Split(getenv("UEM_ALLOWED_COMMANDS", "echo,hostname,uname"), ",") {
		c := strings.TrimSpace(cmd)
		if c != "" {
			cfg.AllowedCommands[c] = struct{}{}
		}
	}
	if len(cfg.AllowedCommands) == 0 {
		return cfg, errors.New("UEM_ALLOWED_COMMANDS must not be empty")
	}
	if cfg.HMACSecret == "" {
		return cfg, errors.New("UEM_HMAC_SECRET is required")
	}
	if !cfg.MTLSDisabled && ((cfg.ClientCertPath == "") != (cfg.ClientKeyPath == "")) {
		return cfg, errors.New("UEM_CLIENT_CERT_PATH and UEM_CLIENT_KEY_PATH must be set together")
	}
	if cfg.EnrollmentToken == "" && (cfg.AgentEmail == "" || cfg.AgentPassword == "") {
		return cfg, errors.New("set UEM_ENROLLMENT_TOKEN or both UEM_AGENT_EMAIL/UEM_AGENT_PASSWORD")
	}
	return cfg, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getenvInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}

func getenvBool(k string, d bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	if v == "" {
		return d
	}
	return v == "1" || v == "true" || v == "yes"
}

func getenvDuration(k string, d time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	x, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return x
}
