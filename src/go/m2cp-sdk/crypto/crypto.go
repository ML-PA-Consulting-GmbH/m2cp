package crypto

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/sha3"
	"io"
	"time"
)

var (
	v1Header         = []byte{0x1}
	v1FixedTimestamp = time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func rsaPublicKey(pubKey *rsa.PublicKey) PublicKey {
	intPubKey := packet.NewRSAPublicKey(v1FixedTimestamp, pubKey)
	return newOpenPGPPubKey(intPubKey)
}

type PublicKey interface {
	ID() string
	Verify(content []byte, sig *packet.Signature) error
	KeyEncoder
}

func encodeDigest(hash crypto.Hash, hashDigest []byte) (string, error) {
	algo := ""
	switch hash {
	case crypto.SHA512:
		algo = "sha512"
	case crypto.SHA3_384:
		algo = "sha3-384"
	default:
		return "", fmt.Errorf("unsupported hash")
	}
	if len(hashDigest) != hash.Size() {
		return "", fmt.Errorf("hash digest by %s should be %d bytes", algo, hash.Size())
	}
	return base64.RawURLEncoding.EncodeToString(hashDigest), nil
}

type KeyEncoder interface {
	keyEncode(w io.Writer) error
}

func encodeKeyBase64(key PublicKey) (string, error) {
	buf := new(bytes.Buffer)
	err := key.keyEncode(buf)
	if err != nil {
		return "", fmt.Errorf("cannot encode key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func newOpenPGPPubKey(intPubKey *packet.PublicKey) *openpgpPubKey {
	h := sha3.New384()
	h.Write(v1Header)
	err := intPubKey.Serialize(h)
	if err != nil {
		panic("internal error: cannot compute public key sha3-384")
	}
	sha3_384, err := encodeDigest(crypto.SHA3_384, h.Sum(nil))
	if err != nil {
		panic("internal error: cannot compute public key sha3-384")
	}
	return &openpgpPubKey{pubKey: intPubKey, sha3_384: sha3_384}
}

type openpgpPubKey struct {
	pubKey   *packet.PublicKey
	sha3_384 string
}

func (opgPubKey *openpgpPubKey) ID() string {
	return opgPubKey.sha3_384
}

func (opgPubKey *openpgpPubKey) Verify(content []byte, sig *packet.Signature) error {
	h := sig.Hash.New()
	h.Write(content)
	return opgPubKey.pubKey.VerifySignature(h, sig)
}

func (opgPubKey openpgpPubKey) keyEncode(w io.Writer) error {
	return opgPubKey.pubKey.Serialize(w)
}
