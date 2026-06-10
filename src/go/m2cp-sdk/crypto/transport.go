package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type CypherEnvelope struct {
	Cyphertext string `json:"cyphertext"`
	Nonce      string `json:"nonce"`
	EncKey     string `json:"encKey"`
}

// EncryptForSecureChannel encrypts the given message using a hybrid encryption scheme.
// IMPORTANT: Only to be used on a secure channel, for example to avoid logging of the plaintext message.
// This implementation does not guarantee authenticity or integrity of the message
// and doesn't protect from replay attacks or rerouting.
func EncryptForSecureChannel(message []byte, pubKeyPEM string) (*CypherEnvelope, error) {
	pubKey, err := parsePEMPublicKey(pubKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM public key: %w", err)
	}

	// Generate random AES key
	aesKey := make([]byte, 32) // 256-bit AES key
	_, err = rand.Read(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate IV
	nonce := make([]byte, gcm.NonceSize()) // Typically 12 bytes for GCM
	_, err = rand.Read(nonce)              // Generate random nonce
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	cyphertext := gcm.Seal(nil, nonce, message, nil)

	enc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	return &CypherEnvelope{
		Cyphertext: base64.StdEncoding.EncodeToString(cyphertext),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		EncKey:     base64.StdEncoding.EncodeToString(enc),
	}, nil

}

func DecryptMessageWithTPM(envelope *CypherEnvelope) ([]byte, error) {
	if envelope == nil {
		return nil, fmt.Errorf("envelope is nil")
	}

	encKeyBytes, err := base64.StdEncoding.DecodeString(envelope.EncKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}
	cyphertext, err := base64.StdEncoding.DecodeString(envelope.Cyphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cyphertext: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	aesKey, err := TpmDecryptRSA(encKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key using TPM decryption key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, cyphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	return plaintext, nil

}

func parsePEMPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an RSA key")
	}
	return rsaPub, nil
}
