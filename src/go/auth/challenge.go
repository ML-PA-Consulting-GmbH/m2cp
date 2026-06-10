package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	v5 "m2cpcli/backend/v5"
	"strings"
)

// GetChallenge queries the server at `url` for a nonce for a user that is known to the server
// by `userEmail` and `sshPublicKey`. The challenge is returned as a byte array.
func GetChallenge(ctx context.Context, url string, userEmail string, sshPublicKey []byte) ([]byte, error) {
	// The URL is passed to the NewGraphqlClient() call via viper!
	viper.Set("store", url)

	// be sure to trim whitespace at the end of the file
	trimmedPublicKey := []byte(strings.TrimSpace(string(sshPublicKey)))

	iulr, err := v5.InitiateUserLogin(ctx, userEmail, string(trimmedPublicKey))
	if err != nil {
		return nil, err
	}
	var challengeString string
	if iulr != nil {
		challengeString = iulr.InitiateUserLogin.Challenge
	} else {
		return nil, fmt.Errorf("failed to initiate user login")
	}

	challengeBytes, err := hex.DecodeString(challengeString)
	if err != nil {
		return nil, fmt.Errorf("could not decode challenge string: %s", err)
	}

	return challengeBytes, nil
}

// SignChallenge cryptographically signs the `challenge` with the users private key.
// If the private key is not password-protected, `password` must be `nil`.
func SignChallenge(challenge, privKeyPem, password []byte) ([]byte, error) {
	var err error
	var signer ssh.Signer

	if password != nil {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privKeyPem, password)
	} else {
		signer, err = ssh.ParsePrivateKey(privKeyPem)
	}
	if err != nil {
		err = fmt.Errorf("could not parse private key: %s", err)
		return []byte(nil), err
	}

	signature, err := signer.Sign(rand.Reader, challenge)
	if err != nil {
		err = fmt.Errorf("could not sign challenge: %s", err)
		return []byte(nil), err
	}

	return signature.Blob, nil
}

// GetJSONWebToken sends the original `challenge` and the corresponding `signedChallenge` to the server at `url`
// to receive a JSON Web Token for authentication.
func GetJSONWebToken(ctx context.Context, url string, challenge []byte, signedChallenge []byte) (string, error) {
	// The URL is passed to the NewGraphqlClient() call via viper!
	viper.Set("store", url)

	// TODO: hex of the bytes? Why not Base64? base64.StdEncoding.EncodeToString()
	challengeHexString := hex.EncodeToString(challenge)
	signedChallengeHexString := hex.EncodeToString(signedChallenge)

	ccr, err := v5.CompleteChallenge(ctx, challengeHexString, signedChallengeHexString)
	if err != nil {
		return "", err
	}
	if ccr == nil || ccr.CompleteChallenge == nil {
		return "", fmt.Errorf("failed to complete challenge")
	}

	return ccr.CompleteChallenge.Token, nil
}
