//go:build windows

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/svc"

	"uem-agent/internal/config"
	"uem-agent/internal/service"
	"uem-agent/internal/transport"
)

const windowsServiceName = "UEM Agent"

func IsWindowsService() bool {
	ok, err := svc.IsWindowsService()
	return err == nil && ok
}

func RunWindowsService(ctx context.Context, cfg config.Config, logger *slog.Logger, client *transport.Client, deviceID string) error {
	agent := service.New(cfg, logger, client, deviceID)
	agent.Run(ctx)
	return nil
}

func HandleServiceCommands() bool {
	if len(os.Args) < 2 {
		return false
	}
	cmd := os.Args[1]
	switch cmd {
	case "install":
		exe, _ := os.Executable()
		exe, _ = filepath.Abs(exe)
		runCmd("sc.exe", "create", windowsServiceName, "binPath=", fmt.Sprintf("\"%s\"", exe), "start=", "auto")
		runCmd("sc.exe", "failure", windowsServiceName, "reset=", "30", "actions=", "restart/5000")
		fmt.Println("Service installed")
		return true
	case "uninstall":
		runCmd("sc.exe", "delete", windowsServiceName)
		fmt.Println("Service uninstalled")
		return true
	case "console":
		return false
	default:
		return false
	}
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("command failed: %v, output=%s\n", err, string(out))
	}
}
