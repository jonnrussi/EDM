//go:build android || ios
// +build android ios

package executor

type Result struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

func Execute(_ string, _ []string, _ map[string]struct{}) Result {
	return Result{
		Success:  false,
		Output:   "local command execution is restricted on mobile agents; use MDM actions",
		ExitCode: 125,
	}
}
