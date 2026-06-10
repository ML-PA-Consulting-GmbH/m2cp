package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PathAbs(path string) (string, error) {
	path = strings.TrimSpace(path)
	if len(path) > 1 && path[0:1] == "~" {
		pathHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot aquire path of user home dir")
		}
		path = filepath.Join(pathHome, path[1:])
	}
	return filepath.Abs(path)
}
