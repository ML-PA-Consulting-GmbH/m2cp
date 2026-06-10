package tools

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeSensorId(t *testing.T) {
	const sensorId = uint32(8667)
	encoded := EncodeSensorId(sensorId)
	if encoded != "2yE" {
		t.Errorf("Encoded sensor id is not correct: %s", encoded)
	}
	decoded, err := DecodeSensorId(encoded)
	assert.NoError(t, err)
	assert.Equal(t, sensorId, decoded)
}

func TestGenerateEncodedSensorIdsForJakob(t *testing.T) {
	ids := []uint32{1, 1000, 8667, 9876, 10001, 100000, 1000000, 10000000, 100000000}
	//ids := []uint32{100000, 1000000, 10000000, 100000000}
	for _, id := range ids {
		encoded := EncodeSensorId(id)
		fmt.Printf("%d %s\n", id, encoded)
		decoded, err := DecodeSensorId(encoded)
		assert.NoError(t, err)
		assert.Equal(t, id, decoded)
	}
}

// TestDecodeSensorId tests decoding of sensor IDs from their encoded string representations.
// Contains padded and unpadded base64 strings.
func TestDecodeSensorId(t *testing.T) {
	tests := []string{
		"AQ",
		"AQA",
		"-Sw",
		"-SwAAA",
		"+Sw",
		"+SwAAA",
	}

	for _, test := range tests {
		decoded, err := DecodeSensorId(test)
		assert.NoError(t, err)
		t.Logf("Decoded sensor id: %d", decoded)
	}
}
