//go:build android || ios
// +build android ios

package executor

import (
	"runtime"
	"time"
)

type Result struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	OS       string `json:"os"`
}

func Execute(_ string, _ []string, _ map[string]struct{}, _ time.Duration) Result {
	return Result{
		Success:  false,
		Output:   "local command execution is restricted on mobile agents; use MDM actions",
		ExitCode: 125,
		OS:       runtime.GOOS,
	}
}
