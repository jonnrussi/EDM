package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"uem-agent/internal/config"
	"uem-agent/internal/executor"
	"uem-agent/internal/inventory"
	"uem-agent/internal/transport"
)

type AgentService struct {
	cfg      config.Config
	logger   *slog.Logger
	client   *transport.Client
	deviceID string
}

type commandEnvelope struct {
	TaskID  string   `json:"task_id"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func New(cfg config.Config, logger *slog.Logger, client *transport.Client, deviceID string) *AgentService {
	return &AgentService{cfg: cfg, logger: logger, client: client, deviceID: deviceID}
}

func (s *AgentService) Run(ctx context.Context) {
	s.logger.Info("agent service started", "device_id", s.deviceID, "poll_interval", s.cfg.PollInterval)
	s.sendInventory(ctx)

	backoff := s.cfg.PollInterval
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("agent shutting down")
			return
		default:
		}

		err := s.pollAndExecute(ctx)
		if err != nil {
			if backoff < s.cfg.RetryMaxDelay {
				backoff *= 2
				if backoff > s.cfg.RetryMaxDelay {
					backoff = s.cfg.RetryMaxDelay
				}
			}
			s.logger.Warn("poll failed", "error", err, "next_attempt_in", backoff)
			time.Sleep(backoff)
			continue
		}

		backoff = s.cfg.PollInterval
		time.Sleep(s.cfg.PollInterval)
	}
}

func (s *AgentService) sendInventory(ctx context.Context) {
	snapshot := inventory.Collect()
	ctxReq, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()
	if err := s.client.DoJSON(ctxReq, transport.Request{Method: "POST", URL: s.cfg.DeviceURL, Body: snapshot, UseAuth: true}, nil); err != nil {
		s.logger.Warn("inventory send failed", "error", err)
	}
}

func (s *AgentService) pollAndExecute(ctx context.Context) error {
	var cmd commandEnvelope
	url := fmt.Sprintf("%s/v1/agent/commands/next?device_id=%s", s.cfg.TaskServiceURL, s.deviceID)
	ctxReq, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()
	if err := s.client.DoJSON(ctxReq, transport.Request{Method: "GET", URL: url, UseAuth: true}, &cmd); err != nil {
		return err
	}
	if cmd.TaskID == "" {
		return nil
	}

	result := executor.Execute(cmd.Command, cmd.Args, s.cfg.AllowedCommands, s.cfg.ExecutionTimeout)
	statusPayload := map[string]any{
		"success":   result.Success,
		"output":    result.Output,
		"exit_code": result.ExitCode,
		"os":        result.OS,
	}
	reportURL := fmt.Sprintf("%s/v1/agent/commands/%s/status", s.cfg.TaskServiceURL, cmd.TaskID)
	ctxReport, cancelReport := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancelReport()
	if err := s.client.DoJSON(ctxReport, transport.Request{Method: "POST", URL: reportURL, Body: statusPayload, UseAuth: true}, nil); err != nil {
		return err
	}
	s.logger.Info("command executed", "task_id", cmd.TaskID, "success", result.Success, "exit_code", result.ExitCode)
	return nil
}
