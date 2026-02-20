//go:build !windows

package main

import (
	"context"
	"log/slog"

	"uem-agent/internal/config"
	"uem-agent/internal/service"
	"uem-agent/internal/transport"
)

func IsWindowsService() bool { return false }

func RunWindowsService(ctx context.Context, cfg config.Config, logger *slog.Logger, client *transport.Client, deviceID string) error {
	agent := service.New(cfg, logger, client, deviceID)
	agent.Run(ctx)
	return nil
}

func HandleServiceCommands() bool { return false }
