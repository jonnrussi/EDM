//go:build windows || (linux && !android) || (darwin && !ios)
// +build windows linux,!android darwin,!ios

package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Result struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

func Execute(command string, args []string, allowed map[string]struct{}) Result {
	if _, ok := allowed[command]; !ok {
		return Result{Success: false, Output: "command not allowed", ExitCode: 126}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{Success: false, Output: "command timeout", ExitCode: 124}
		}
		if ee, ok := err.(*exec.ExitError); ok {
			return Result{Success: false, Output: output, ExitCode: ee.ExitCode()}
		}
		return Result{Success: false, Output: fmt.Sprintf("execution error: %v", err), ExitCode: 1}
	}
	return Result{Success: true, Output: output, ExitCode: 0}
}
