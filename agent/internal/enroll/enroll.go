package enroll

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"uem-agent/internal/transport"
)

type request struct {
	EnrollmentToken string `json:"enrollment_token"`
	Hostname        string `json:"hostname"`
	Platform        string `json:"platform"`
}

type response struct {
	DeviceID    string `json:"device_id"`
	AccessToken string `json:"access_token"`
}

func LoadDeviceID(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func SaveDeviceID(path, deviceID string) error {
	return os.WriteFile(path, []byte(deviceID), 0o600)
}

func EnsureEnrollment(ctx context.Context, client *transport.Client, url, enrollmentToken, hostname, platform, deviceIDPath string) (string, string, error) {
	existingID, err := LoadDeviceID(deviceIDPath)
	if err != nil {
		return "", "", err
	}
	if existingID != "" {
		return existingID, "", nil
	}
	if enrollmentToken == "" {
		return "", "", errors.New("missing UEM_ENROLLMENT_TOKEN for first run")
	}
	var out response
	err = client.DoJSON(ctx, transport.Request{
		Method: "POST",
		URL:    url,
		Body: request{
			EnrollmentToken: enrollmentToken,
			Hostname:        hostname,
			Platform:        platform,
		},
		UseAuth: false,
	}, &out)
	if err != nil {
		return "", "", err
	}
	if out.DeviceID == "" {
		return "", "", errors.New("empty device_id from enrollment")
	}
	if err := SaveDeviceID(deviceIDPath, out.DeviceID); err != nil {
		return "", "", err
	}
	return out.DeviceID, out.AccessToken, nil
}

func ReadRaw(path string) (json.RawMessage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}
