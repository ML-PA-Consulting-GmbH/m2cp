package helper

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"regexp"
	"strings"
)

func IsValidColonSeparatedHexadecimalEncodedBytes(input string) bool {
	// Remove all whitespace (spaces, tabs, newlines, etc.)
	cleaned := strings.Join(strings.Fields(input), "")

	// Regular expression to match exactly 32 hex byte pairs separated by colons
	// ^(?:[0-9a-fA-F]{2}:){31}[0-9a-fA-F]{2}$  ensures 32 bytes, 31 colons
	pattern := `^(?:[0-9a-fA-F]{2}:){31}[0-9a-fA-F]{2}$`

	matched, err := regexp.MatchString(pattern, cleaned)
	if err != nil {
		return false
	}
	return matched
}

// Convert colon-separated hex string to []byte
func parseHexString(input string) ([]byte, error) {
	// Remove all whitespace (spaces, tabs, newlines, etc.)
	cleaned := strings.Join(strings.Fields(input), "")
	parts := strings.Split(cleaned, ":")

	if len(parts) != 32 {
		return nil, fmt.Errorf("expected 32 bytes, got %d", len(parts))
	}

	key := make([]byte, 32)
	for i, part := range parts {
		b, err := hex.DecodeString(part)
		if err != nil || len(b) != 1 {
			return nil, fmt.Errorf("invalid hex byte: %q", part)
		}
		key[i] = b[0]
	}
	return key, nil
}

// Convert raw Ed25519 private key bytes to PEM
func ed25519PrivateKeyToPEM(privateKeyBytes []byte) ([]byte, error) {
	if len(privateKeyBytes) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid Ed25519 seed length: %d", len(privateKeyBytes))
	}

	// Create full Ed25519 private key from seed
	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)

	// Convert to PKCS#8 DER
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %v", err)
	}

	// Encode to PEM
	var pemBuf bytes.Buffer
	err = pem.Encode(&pemBuf, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode PEM: %v", err)
	}

	return pemBuf.Bytes(), nil
}

func ColonSeparatedHexadecimalEncodedBytesToPemFormatedPrivateKeyBytes(input string) ([]byte, error) {
	seed, err := parseHexString(input)
	if err != nil {
		return nil, err
	}

	pemStr, err := ed25519PrivateKeyToPEM(seed)
	if err != nil {
		return nil, err
	}
	return pemStr, nil
}

func Sha256Fingerprint(pub ed25519.PublicKey) string {
	hash := sha256.Sum256(pub)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func LoadEd25519PrivateKeyFromPEM(pemBytes []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an Ed25519 private key")
	}

	return edKey, nil
}
