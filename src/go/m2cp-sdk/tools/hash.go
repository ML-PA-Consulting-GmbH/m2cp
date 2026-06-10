package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func Sha256Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)
	// encode with hex
	return hex.EncodeToString(hash), nil
	//return base64.StdEncoding.EncodeToString(hash), nil
}
