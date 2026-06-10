package tools

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ExecuteCommand executes a shell command and returns its output or an error.
func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	return out.String(), nil
}
