package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"uem-agent/internal/auth"
	"uem-agent/internal/config"
	"uem-agent/internal/enroll"
	"uem-agent/internal/platform"
	"uem-agent/internal/service"
	"uem-agent/internal/transport"
)

func main() {
	if HandleServiceCommands() {
		return
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(rootCtx, cfg, logger); err != nil {
		logger.Error("agent terminated with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	client, err := transport.NewClient(
		cfg.ClientCertPath,
		cfg.ClientKeyPath,
		cfg.CACertPath,
		cfg.HMACSecret,
		cfg.MTLSDisabled,
		cfg.RequestTimeout,
		cfg.RetryMaxAttempts,
		cfg.RetryBaseDelay,
		cfg.RetryMaxDelay,
	)
	if err != nil {
		return err
	}

	var tm *auth.TokenManager
	if cfg.AgentEmail != "" && cfg.AgentPassword != "" {
		tm = auth.NewTokenManager(client, cfg.AuthURL, cfg.AgentEmail, cfg.AgentPassword, cfg.RefreshLeeway)
		client.SetTokenProvider(tm.Token)
		client.SetUnauthorizedHandler(tm.Login)
		if err := tm.Login(ctx); err != nil {
			logger.Warn("initial login failed", "error", err)
		}
		tm.Start(ctx)
	}

	deviceID, bootstrapToken, err := enroll.EnsureEnrollment(
		ctx,
		client,
		cfg.EnrollmentURL,
		cfg.EnrollmentToken,
		platform.Hostname(),
		platform.Name(),
		cfg.DeviceIDFile,
	)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if deviceID == "" {
		if d, loadErr := enroll.LoadDeviceID(cfg.DeviceIDFile); loadErr == nil && d != "" {
			deviceID = d
		}
	}
	if deviceID == "" {
		return errors.New("device_id unavailable after enrollment")
	}

	if bootstrapToken != "" {
		client.SetTokenProvider(func() string {
			if tm != nil && tm.Token() != "" {
				return tm.Token()
			}
			return bootstrapToken
		})
	}

	if IsWindowsService() {
		return RunWindowsService(ctx, cfg, logger, client, deviceID)
	}

	agent := service.New(cfg, logger, client, deviceID)
	agent.Run(ctx)
	return nil
}
