package inventory

import (
	"os"
	"runtime"
)

type Snapshot struct {
	Hostname string `json:"hostname"`
	OSName   string `json:"os_name"`
	OSVer    string `json:"os_version"`
	CPUArch  string `json:"cpu"`
	RAMMB    int    `json:"ram_mb"`
}

func Collect() Snapshot {
	host, _ := os.Hostname()
	return Snapshot{
		Hostname: host,
		OSName:   runtime.GOOS,
		OSVer:    "unknown",
		CPUArch:  runtime.GOARCH,
		RAMMB:    4096,
	}
}
