package platform

import (
	"os"
	"runtime"
)

func Hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func Name() string {
	return runtime.GOOS
}
