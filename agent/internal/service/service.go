package service

import (
	"log"
	"strings"
	"time"

	"uem-agent/internal/executor"
	"uem-agent/internal/inventory"
	"uem-agent/internal/transport"
)

type Config struct {
	DeviceID        string
	DeviceURL       string
	TaskServiceURL  string
	PollIntervalSec int
	AllowedCommands string
}

type AgentService struct {
	cfg     Config
	client  *transport.Client
	allowed map[string]struct{}
}

type commandEnvelope struct {
	TaskID  string   `json:"task_id"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func New(cfg Config, client *transport.Client) *AgentService {
	allowed := map[string]struct{}{}
	for _, c := range strings.Split(cfg.AllowedCommands, ",") {
		trimmed := strings.TrimSpace(c)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		allowed["echo"] = struct{}{}
		allowed["uname"] = struct{}{}
		allowed["hostname"] = struct{}{}
	}
	return &AgentService{cfg: cfg, client: client, allowed: allowed}
}

func (s *AgentService) Run() {
	pollInterval := time.Duration(s.cfg.PollIntervalSec) * time.Second
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Printf("agent service started; device_id=%s, poll_interval=%s", s.cfg.DeviceID, pollInterval)
	s.sendInventory()
	for {
		s.pollAndExecute()
		<-ticker.C
	}
}

func (s *AgentService) sendInventory() {
	snapshot := inventory.Collect()
	if err := s.client.DoJSON("POST", s.cfg.DeviceURL, snapshot, nil); err != nil {
		log.Printf("inventory send failed: %v", err)
	}
}

func (s *AgentService) pollAndExecute() {
	var cmd commandEnvelope
	url := s.cfg.TaskServiceURL + "/v1/agent/commands/next?device_id=" + s.cfg.DeviceID
	if err := s.client.DoJSON("GET", url, nil, &cmd); err != nil {
		log.Printf("poll failed: %v", err)
		return
	}
	if cmd.TaskID == "" {
		return
	}

	result := executor.Execute(cmd.Command, cmd.Args, s.allowed)
	statusPayload := map[string]any{
		"success":   result.Success,
		"output":    result.Output,
		"exit_code": result.ExitCode,
	}
	reportURL := s.cfg.TaskServiceURL + "/v1/agent/commands/" + cmd.TaskID + "/status"
	if err := s.client.DoJSON("POST", reportURL, statusPayload, nil); err != nil {
		log.Printf("status report failed for task %s: %v", cmd.TaskID, err)
	}
}
