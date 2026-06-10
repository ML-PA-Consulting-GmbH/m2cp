package tools

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
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

func DecodeSensorId(encodedID string) (uint32, error) {

	decodedBytes, errURL := base64.RawURLEncoding.DecodeString(encodedID)
	if errURL != nil {
		var errStd error
		if decodedBytes, errStd = base64.RawStdEncoding.DecodeString(encodedID); errStd != nil {
			return 0, errors.New("failed to decode base64:\n" + "- for URL safe encoding:" + errURL.Error() + "\n" + "- for standard encoding" + errStd.Error())
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

func EncodeSensorId(id uint32) string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, id)

	// Trim trailing zero bytes to handle numbers that don't use all 4 bytes
	i := len(buf)
	for i > 0 && buf[i-1] == 0 {
		i--
	}
	if i != len(buf) {
		buf = buf[:i]
	}

	return base64.RawStdEncoding.EncodeToString(buf)
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
