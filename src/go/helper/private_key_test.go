package helper

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestIsValidColonSeparatedHexadecimalEncodedBytes(t *testing.T) {
	defer goleak.VerifyNone(t)

	withWhitespace := `15:a4:15:12:ab:7b:70:ca:53:e7:9e:6a:2d:54:3d:
    58:04:6d:64:c9:53:7b:a5:af:7c:65:9f:97:b1:b1:
    49:aa`
	assert.True(t, IsValidColonSeparatedHexadecimalEncodedBytes(withWhitespace))
}

func TestColonSeparatedHexadecimalEncodedBytesToPemFormatedPrivateKeyBytes(t *testing.T) {
	defer goleak.VerifyNone(t)

	encodedKey := `15:a4:15:12:ab:7b:70:ca:53:e7:9e:6a:2d:54:3d:
    58:04:6d:64:c9:53:7b:a5:af:7c:65:9f:97:b1:b1:
    49:aa`
	pem, err := ColonSeparatedHexadecimalEncodedBytesToPemFormatedPrivateKeyBytes(encodedKey)
	assert.NoError(t, err)
	assert.NotNil(t, pem)

	assert.Equal(t, `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBWkFRKre3DKU+eeai1UPVgEbWTJU3ulr3xln5exsUmq
-----END PRIVATE KEY-----
`, string(pem))

}
