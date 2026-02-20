//go:build windows || (linux && !android) || (darwin && !ios)
// +build windows linux,!android darwin,!ios

package executor

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Result struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	OS       string `json:"os"`
}

func Execute(command string, args []string, allowed map[string]struct{}, timeout time.Duration) Result {
	if _, ok := allowed[command]; !ok {
		return Result{Success: false, Output: "command not allowed", ExitCode: 126, OS: runtime.GOOS}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return Result{Success: false, Output: "command timeout", ExitCode: 124, OS: runtime.GOOS}
		}
		if ee, ok := err.(*exec.ExitError); ok {
			return Result{Success: false, Output: output, ExitCode: ee.ExitCode(), OS: runtime.GOOS}
		}
		return Result{Success: false, Output: fmt.Sprintf("execution error: %v", err), ExitCode: 1, OS: runtime.GOOS}
	}
	return Result{Success: true, Output: output, ExitCode: 0, OS: runtime.GOOS}
}
