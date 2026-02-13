package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/dreams-money/opnsense-failover/config"
)

func GetLeaderName(cfg config.Config) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "linux" || isCygwin() {
		cmd = exec.Command("bash", "-c", cfg.LeaderScript)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("CMD.exe", "/C", cfg.LeaderScript)
	} else {
		return "", fmt.Errorf("Unknown Runtime - OS: %v, TERM: %v", runtime.GOOS, os.Getenv("TERM"))
	}

	output, err := cmd.Output()

	str := string(output)
	str = strings.TrimSpace(str)

	return str, err
}

func isCygwin() bool {
	return os.Getenv("TERM") == "xterm"
}
