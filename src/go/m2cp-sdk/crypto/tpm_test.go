package crypto

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestTpmEncodePubKey(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping tpm tests in CI environment")
	}

	fmt.Printf("foo")
	ekPub, err := TpmGetEndorsementPublicKey()
	assert.NoError(t, err)

	// doesn't work as crypto.PublicKey can't implement /snapcore/snapd/asserts.PublicKey as it has a non-exported method and is defined in a different package
	//pubEncoded, err := asserts.EncodePublicKey(ekPub)
	pubEncoded, err := encodeKeyBase64(ekPub)
	assert.NoError(t, err)

	fmt.Printf("X-Tpm-Ek: %s\n", string(pubEncoded))
	fmt.Printf("X-Tpm-Ek-Sha3-384: %s\n", ekPub.ID())
}

func TestTpmSignBytes(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping tpm tests in CI environment")
	}

	signature, err := TpmSignBytes([]byte("hello world"))
	assert.NoError(t, err)
	assert.NotNil(t, signature)
}

func TestTpmGetEndorsementPublicKey(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping tpm tests in CI environment")
	}

	pub, err := TpmGetEndorsementPublicKey()
	assert.NoError(t, err)
	assert.NotNil(t, pub)
}

func TestTpmPushEkWithHeader(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping tpm tests in CI environment")
	}

	ekPub, err := TpmGetEndorsementPublicKey()
	assert.NoError(t, err)
	assert.NotNil(t, ekPub)

	pubEncoded, err := encodeKeyBase64(ekPub)
	assert.NoError(t, err)
	assert.NotNil(t, pubEncoded)

	fmt.Printf("X-Tpm-Ek: %s\n", string(pubEncoded))
}

func TestGetDriveSerial(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping tpm tests in CI environment")
	}

	serial, err := getDriveSerial("/dev/sda")
	fmt.Println(serial)
	assert.NoError(t, err)
	assert.NotEmpty(t, serial)
}

func TestGetMacAddresses(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping tpm tests in CI environment")
	}

	macs, err := getMacAddresses()
	assert.NoError(t, err)
	assert.NotEmpty(t, macs)
	fmt.Println(macs)
	fmt.Println(err)
}
