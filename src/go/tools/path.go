package tools

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
)

func Abspath(path string) (string, error) {
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

func ReadLocalFile(path string) ([]byte, error) {
	path, err := Abspath(path)
	if err != nil {
		return nil, fmt.Errorf("could not derive absolute path from '%s'", path)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file '%s'", path)
	}
	return contents, nil
}

func MakeTempFileName() string {
	return "/tmp/" + uuid.New().String()
}
