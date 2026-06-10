package suit

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

const exampleTar = "../test/data/m2cp-fw-dft-sensor_3.2.0_arm32.tar"

func TestIsValidFirmwareTar(t *testing.T) {
	defer goleak.VerifyNone(t)

	err := ValidateFirmwareTar(exampleTar)
	assert.NoError(t, err)
}

func TestExtractFirmwareMetadataFromTar(t *testing.T) {
	defer goleak.VerifyNone(t)

	fw, err := ExtractFirmwareMetadataFromTar(exampleTar)
	assert.NoError(t, err)
	assert.NotNil(t, fw)

	assert.NotEqual(t, "error", fw.String())
	//fmt.Println(fw.String())
	assert.Equal(t, "{\"app-name\":\"dft-sensor\",\"app-summary\":\"DFT Sensor board firmware for the refresh 2.0 \",\"app-version\":\"3.2.0\",\"device-model\":\"dft_sensor\",\"architecture\":\"arm32\",\"class-id\":\"\",\"bootable\":false,\"offset\":null}", fw.String())
}

func TestValidateFirmwareBlob(t *testing.T) {
	defer goleak.VerifyNone(t)

	err := ValidateFirmwareBlobFromTar(exampleTar, "./fw0.bin")
	assert.NoError(t, err)

	err = ValidateFirmwareBlobFromTar(exampleTar, "./fw1.bin")
	assert.NoError(t, err)
}
