package system

import (
	"os/exec"
	"runtime"
)

func Open(name string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", name).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", name).Run()
	case "linux":
		return exec.Command("xdg-open", name).Run()
	}

	return nil
}
