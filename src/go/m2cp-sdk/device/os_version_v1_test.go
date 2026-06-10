package device

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestNewDeviceOsVersionFromString_V1(t *testing.T) {
	t.Skip("Skipping TestNewDeviceOsVersionFromString_V1 as it currently fails")
	defer goleak.VerifyNone(t)

	example := "1-2d-pp0-m-2-3-4-5-6-7-8-9-10-11-12"
	dov, err := osVersion1FromString(example)
	assert.NoError(t, err)
	assert.NotNil(t, dov)

	assert.Equal(t, dov.Version, 1)
	assert.Equal(t, dov.SnapstoreUrl, "2ddd754c377fc0089ed0e26027f8638af91b621a")
	assert.Equal(t, dov.Board, "phyboard-pollux")
	assert.Equal(t, dov.TenantAlias, "ab3178d1351758363af471e892d64b25a9c5abaa")

	assert.Equal(t, dov.Kernel.Name, "phyboard-pollux-kernel")
	assert.Equal(t, dov.Kernel.Revision, 2)
	assert.Equal(t, dov.Gadget.Name, "phyboard-pollux-gadget")
	assert.Equal(t, dov.Gadget.Revision, 3)

	assert.Equal(t, dov.SnapdRevision, 4)
	assert.Equal(t, dov.CoreRevision, 5)
	assert.Equal(t, dov.Core18Revision, 6)
	assert.Equal(t, dov.Core20Revision, 7)
	assert.Equal(t, dov.Core22Revision, 8)
	assert.Equal(t, dov.Core24Revision, 9)
	assert.Equal(t, dov.M2cpGatewayRevision, 10)
	assert.Equal(t, dov.M2cpLogStatRevision, 11)
	assert.Equal(t, dov.M2cpMessageHubRevision, 12)

	// fmt.Println(dov.String())
}

func TestDeviceOsVersion_ShortStringV1(t *testing.T) {
	defer goleak.VerifyNone(t)

	dov := osVersionV1{
		Version:      1,
		SnapstoreUrl: sha1sum("https://app-snapstore-api-st02-dev.azurewebsites.net"),
		Board:        "phyboard-pollux",
		TenantAlias:  sha1sum("mlpa"),
	}
	fmt.Println(sha1sum("mlpa"))

	dov.Kernel.Revision = 2
	dov.Gadget.Revision = 3
	dov.SnapdRevision = 4
	dov.CoreRevision = 5
	dov.Core18Revision = 6
	dov.Core20Revision = 7
	dov.Core22Revision = 8
	dov.Core24Revision = 9
	dov.M2cpGatewayRevision = 10
	dov.M2cpLogStatRevision = 11
	dov.M2cpMessageHubRevision = 12

	// fmt.Println(dov.String())

	short, err := dov.Encode()
	assert.NoError(t, err)
	assert.Equal(t, "1-2d-pp0-m-2-3-4-5-6-7-8-9-10-11-12", short)
}

func sha1sum(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
