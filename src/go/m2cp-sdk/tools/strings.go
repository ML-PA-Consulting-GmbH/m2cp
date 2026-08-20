package tools

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// IsPrintable checks if the given byte slice is composed of printable characters.
func IsPrintable(b []byte) bool {
	for _, c := range b {
		if !unicode.IsPrint(rune(c)) && !unicode.IsSpace(rune(c)) {
			return false
		}
	}
	return true
}

// FormatHex formats a byte slice into a hex string with spaces and fixed blocks per line.
func FormatHex(b []byte) string {
	hexStr := hex.EncodeToString(b)
	var sb strings.Builder

	blockSize := 2 // Size of each hex block
	blocksPerLine := 16
	for i, c := range hexStr {
		sb.WriteRune(c)
		if i%blockSize == blockSize-1 {
			sb.WriteRune(' ')
		}
		if (i+1)%(blockSize*blocksPerLine) == 0 && i != len(hexStr)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func trimTrailingZeros(buf []byte) []byte {
	i := len(buf)
	for i > 0 && buf[i-1] == 0 {
		i--
	}
	if i != len(buf) {
		buf = buf[:i]
	}
	return buf
}

func encodeHardwareSerial(hardwareSerial uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, hardwareSerial)
	return buf
}

// DecodeHardwareSerial transforms encoded Hardware Serial to an integer value. The hardware serial is a
// 32-bit unsigned integer in little-endian byte order, encoded as Base 64 using the
// URL and filename safe alphabet (see also: https://datatracker.ietf.org/doc/html/rfc4648#section-5). It must not have
// trailing pad characters (`=`) and may have trimmed zero bytes.
func DecodeHardwareSerial(encodedHardwareSerial string) (uint32, error) {
	decodedBytes, errURL := base64.RawURLEncoding.DecodeString(encodedHardwareSerial)
	if errURL != nil {
		var errStd error
		if decodedBytes, errStd = base64.RawStdEncoding.DecodeString(encodedHardwareSerial); errStd != nil {
			return 0, fmt.Errorf("failed to decode base64: URL safe encoding failed with: %w: standard encoding failed with: %w", errURL, errStd)
		}
	}

	// right pad with zero bytes to handle numbers that don't use all 4 bytes
	if len(decodedBytes) < 4 {
		decodedBytes = append(decodedBytes, make([]byte, 4-len(decodedBytes))...)
	}

	switch len(decodedBytes) {
	//case 1:
	//	return uint32(decodedBytes[0]), nil
	//case 2:
	//	return uint32(binary.LittleEndian.Uint16(decodedBytes)), nil
	//case 3:
	//	decodedBytes = append(decodedBytes, 0)
	//	return uint32(binary.LittleEndian.Uint32(decodedBytes)), nil
	case 4:
		return uint32(binary.LittleEndian.Uint32(decodedBytes)), nil
	case 8:
		return binary.LittleEndian.Uint32(decodedBytes), nil
	default:
		return 0, errors.New("invalid length of decoded byte slice")
	}
}

// Deprecated: use DecodeHardwareSerial instead.
//
//go:fix inline
func DecodeSensorId(encodedID string) (uint32, error) {
	return DecodeHardwareSerial(encodedID)
}

// EncodeHardwareSerial transforms the Hardware Serial to the encoded representation as a string. The hardware serial is a
// 32-bit unsigned integer in little-endian byte order, encoded as Base 64 using the
// URL and filename safe alphabet (see also: https://datatracker.ietf.org/doc/html/rfc4648#section-5). It must not have
// trailing pad characters (`=`) and zero bytes are not trimmed.
func EncodeHardwareSerial(hardwareSerial uint32) string {
	return base64.RawURLEncoding.EncodeToString(encodeHardwareSerial(hardwareSerial))
}

// Deprecated: use EncodeHardwareSerial instead.
//
//go:fix inline
func EncodeSensorId(id uint32) string {
	return EncodeHardwareSerial(id)
}

// Deprecated: use EncodeHardwareSerial without trimming instead.
// EncodeSensorIdTrimmed returns the encoded hardware serial without trailing zero bytes.
func EncodeSensorIdTrimmed(id uint32) string {
	return base64.RawURLEncoding.EncodeToString(trimTrailingZeros(encodeHardwareSerial(id)))
}

func ConvertToISO8601UTC(timeStr string) (string, error) {
	// Parse the time string to a time.Time object.
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return "", err
	}
	return FormatISO8601UTC(t), nil
}

func FormatISO8601UTC(t time.Time) string {
	t = t.UTC()
	return t.Format("20060102T150405Z")
}

func StrPtr(s string) *string {
	return &s
}
