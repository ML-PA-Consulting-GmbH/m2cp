package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"m2cpcli/tools"
)

const (
	EnterSshPasswordUsingStdin string = "ENTER_SSH_USING_STDIN"
)

func ReadLocalPrivateSshKey(path string, passphrase string) (ssh.Signer, error) {
	pem, err := tools.ReadLocalFile(path)
	if err != nil {
		return nil, err
	}
	var signer ssh.Signer
	if passphrase == "" || passphrase == EnterSshPasswordUsingStdin {
		signer, err = ssh.ParsePrivateKey(pem)
	}
	if (signer == nil) ||
		(err != nil && err.Error() == "ssh: this private key is passphrase protected") {
		if passphrase == EnterSshPasswordUsingStdin {
			passphrase, err = tools.ConsoleInputPassword(fmt.Sprintf("Enter passphrase for ssh key '%s':", path))
			if err != nil {
				return nil, err
			}
		}
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pem, []byte(passphrase))
	}
	if err != nil {
		return nil, fmt.Errorf("error reading ssh private key (using password: %v)", passphrase != "")
	}
	return signer, nil
}

func ReadLocalPublicSshKey(path string) (ssh.PublicKey, error) {
	pem, err := tools.ReadLocalFile(path)
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pem)
	return pubKey, err
}

func SignNonce(nonce []byte, account string, storeName string, signer ssh.Signer) ([]byte, error) {
	signature, err := signer.Sign(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed signing nonce with users ssh key")
	}

	signatureSerialized, err := json.Marshal(SignedNonce{
		Account:   account,
		StoreName: storeName,
		Nonce:     string(nonce),
		Format:    signature.Format,
		Signature: base64.StdEncoding.EncodeToString(signature.Blob),
	})
	if err != nil {
		return nil, err
	}
	return signatureSerialized, err
}

func VerifySignedNonce(signedNonce SignedNonce, pem []byte) error {
	var err error
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pem)
	if err != nil {
		return err
	}
	blobDecoded, err := base64.StdEncoding.DecodeString(signedNonce.Signature)
	if err != nil {
		return err
	}
	sig := &ssh.Signature{
		Format: signedNonce.Format,
		Blob:   blobDecoded,
	}
	err = pubKey.Verify([]byte(signedNonce.Nonce), sig)
	if err != nil {
		return err
	}
	return nil
}
