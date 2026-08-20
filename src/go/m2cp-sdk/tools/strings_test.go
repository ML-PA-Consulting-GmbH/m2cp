package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecodeHardwareSerial(t *testing.T) {
	var hardwareSerials = map[uint32]string{
		uint32(1):         "AQAAAA",
		uint32(1000):      "6AMAAA",
		uint32(100000):    "oIYBAA",
		uint32(1000000):   "QEIPAA",
		uint32(10000000):  "gJaYAA",
		uint32(100000000): "AOH1BQ",
		uint32(10001):     "EScAAA",
		uint32(10235):     "-ycAAA",
		uint32(10236):     "_CcAAA",
		uint32(8667):      "2yEAAA",
		uint32(9876):      "lCYAAA",
	}
	for hardwareSerial, expectedEncoding := range hardwareSerials {
		actualEncoding := EncodeHardwareSerial(hardwareSerial)
		assert.Equalf(t, expectedEncoding, actualEncoding, "EncodeHardwareSerial(%d) failed: %s != %s",
			hardwareSerial, actualEncoding, expectedEncoding)

		decoded, err := DecodeHardwareSerial(actualEncoding)
		assert.NoError(t, err)
		assert.Equal(t, hardwareSerial, decoded)
	}
}

func TestEncodeSensorIdTrimmed(t *testing.T) {
	var hardwareSerials = map[uint32]string{
		uint32(1):         "AQ",
		uint32(1000):      "6AM",
		uint32(100000):    "oIYB",
		uint32(1000000):   "QEIP",
		uint32(10000000):  "gJaY",
		uint32(100000000): "AOH1BQ",
		uint32(10001):     "ESc",
		uint32(10235):     "-yc",
		uint32(10236):     "_Cc",
		uint32(8667):      "2yE",
		uint32(9876):      "lCY",
	}
	for hardwareSerial, expectedEncoding := range hardwareSerials {
		actualEncoding := EncodeSensorIdTrimmed(hardwareSerial)
		assert.Equalf(t, expectedEncoding, actualEncoding, "EncodeSensorIdTrimmed(%d) failed: %s != %s",
			hardwareSerial, actualEncoding, expectedEncoding)

		decoded, err := DecodeHardwareSerial(actualEncoding)
		assert.NoError(t, err)
		assert.Equal(t, hardwareSerial, decoded)
	}
}

func TestDecodeHardwareSerial(t *testing.T) {
	var encodedHardwareSerials = map[string]uint32{
		"AQ":     uint32(1),
		"AQA":    uint32(1),
		"-Sw":    uint32(11513),
		"-SwAAA": uint32(11513),
		"+Sw":    uint32(11513),
		"+SwAAA": uint32(11513),
	}

	for encodedValue, expectedDecoded := range encodedHardwareSerials {
		actualDecoded, err := DecodeHardwareSerial(encodedValue)
		assert.NoError(t, err)
		assert.Equal(t, expectedDecoded, actualDecoded)
		// t.Logf("Decoded sensor id: %d", actualDecoded)
	}
}
