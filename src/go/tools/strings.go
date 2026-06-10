package tools

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"m2cpcli/structs"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
)

func StreamToString(r io.Reader) string {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, r)
	if err != nil {
		return ""
	}
	return fmt.Sprint(buf.String())
}

func StreamToBytes(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(stream)
	return buf.Bytes()
}

func MaybeStringToString(maybe *string, defaultValue string) string {
	if maybe != nil {
		return *maybe
	}
	return defaultValue
}

func MaybeTimeToString(maybe *time.Time, format, defaultValue string) string {
	if maybe == nil {
		return defaultValue
	}
	return maybe.Format(format)
}

func StringToMaybeTime(s string, format string) *time.Time {
	parsed, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &parsed
	}
	return nil
}

func BoolToString(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}

func MaybeBoolToString(maybe *bool, yes, no, undefined string) string {
	if maybe == nil {
		return undefined
	} else if *maybe == true {
		return yes
	}
	return no
}

func MaybeIntToString(maybe *int, defaultValue string) string {
	if maybe != nil {
		return fmt.Sprintf("%d", *maybe)
	}
	return defaultValue
}

func MaybeInt64ToString(maybe *int64, defaultValue string) string {
	if maybe != nil {
		return fmt.Sprintf("%d", *maybe)
	}
	return defaultValue
}

func MaybeInt64ToStringFormat(maybe *int64, defaultValue string, formatter func(int64) string) string {
	if maybe != nil {
		return formatter(*maybe)
	}
	return defaultValue
}

func MaybeFlagToMaybeString(cmd *cobra.Command, flag string) *string {
	if cmd.Flags().Changed(flag) {
		if val, err := cmd.Flags().GetString(flag); err != nil {
			return nil
		} else {
			return &val
		}
	}
	return nil
}

func FormatFileSize(n int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case n >= GB:
		return fmt.Sprintf("%.1fGB", float64(n)/GB)
	case n >= MB:
		return fmt.Sprintf("%.1fMB", float64(n)/MB)
	case n >= KB:
		return fmt.Sprintf("%.1fKB", float64(n)/KB)
	default:
		return fmt.Sprintf("%dB", n)
	}
}

func FormatSeconds(seconds int64) string {
	// Calculate days, hours, minutes, and seconds
	days := seconds / (24 * 3600)
	seconds = seconds % (24 * 3600)
	hours := seconds / 3600
	seconds = seconds % 3600
	minutes := seconds / 60
	seconds = seconds % 60

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%ds ", seconds)
	}

	return result
}

func CompactJsonString(data string) string {
	compactedBuffer := new(bytes.Buffer)
	err := json.Compact(compactedBuffer, []byte(data))
	if err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("%v", compactedBuffer)
}

func JsonToStruct(jsonstr []byte, p interface{}) error {
	return json.Unmarshal(jsonstr, p)
}

func ListContainsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func StringArrayEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func StringArraysContainEqualElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, aItem := range a {
		found := false
		for _, bItem := range b {
			if aItem == bItem {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func StringListToJson(list []string) string {
	return `["` + strings.Join(list[:], `" , "`) + `"]`
}

func RandomDigits(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	out := ""
	for i := 0; i < length; i++ {
		n := r.Intn(10)
		out = fmt.Sprintf("%s%d", out, n)
	}
	return out
}

func ConsoleInput(message string) string {
	if message != "" {
		fmt.Printf("%s ", message)
	}
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	// convert CRLF to LF
	text = strings.Replace(text, "\n", "", -1)
	return text
}

func ConsoleInputPassword(message string) (string, error) {
	if message != "" {
		fmt.Printf("%s ", message)
	}
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return "", err
	}
	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}

func StringSha3384(input string) string {
	h := sha3.New384()
	h.Write([]byte(input))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func IsValidAppId(snapId string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{32}$")
	if r.MatchString(snapId) {
		withDashes := fmt.Sprintf("%s-%s-%s-%s-%s", snapId[0:0+8], snapId[8:8+4], snapId[12:12+4],
			snapId[16:16+4], snapId[20:20+12])
		return IsValidUuid(withDashes)
	}
	return false
}

// IsValidAppName checks if a given Snap name is valid.
// Snap names can only use ASCII lowercase letters, numbers, and hyphens; they must have at least one letter;
// they cannot start or end with a hyphen; and they  cannot have two hyphens in a row.
func IsValidAppName(snapName string) bool {
	generalPattern := regexp.MustCompile("^[a-z0-9-]*[a-z][a-z0-9-]*$")
	if !generalPattern.MatchString(snapName) {
		return false
	}
	if strings.HasPrefix(snapName, "-") || strings.HasSuffix(snapName, "-") {
		return false
	}
	if strings.Contains(snapName, "--") {
		return false
	}
	return true
}

func IsValidSnapVersion(version string) bool {
	// See also: https://regex101.com/r/3T5yAP/1
	r := regexp.MustCompile(`^[0-9]+(\.[0-9]+)*(-[a-zA-Z0-9]+)?$`)
	return r.MatchString(version)
}

func IsValidDeviceArchitecture(snapDeviceArchitecture string) bool {
	return slices.Contains(structs.DeviceArchitectures, structs.DeviceArchitecture(strings.ToUpper(snapDeviceArchitecture)))
}

func IsValidDeviceType(deviceType string) bool {
	_, ok := structs.DeviceTypeAliases[deviceType]
	return ok
}

// CanonicalDeviceTypeFromUserString tries to match what the user means by best effort and returns the string that is
// used for the DeviceType in the backend.
func CanonicalDeviceTypeFromUserString(userMeantDeviceType string) (result string, ok bool) {
	result = userMeantDeviceType
	ok = false

	// remove whitespaces and dashes
	re := regexp.MustCompile(`[\s-]+`)
	result = re.ReplaceAllString(result, "")

	result = strings.ToLower(result)

	// TODO: this is similar to legacy.deviceTypeAliases. But that's hidden. And not equal.
	switch {
	case result == "rtd":
		result = string(structs.DeviceTypeRealTimeDevice)
		ok = true
	case strings.Contains(result, "realtime"):
		result = string(structs.DeviceTypeRealTimeDevice)
		ok = true
	case result == "ed":
		result = string(structs.DeviceTypeEdgeDevice)
		ok = true
	case strings.Contains(result, "edge"):
		result = string(structs.DeviceTypeEdgeDevice)
		ok = true
	}

	return result, ok
}

func IsValidRevision(snapRevision string) bool {
	// Either 'latest' or a number
	r := regexp.MustCompile(`^(latest|[0-9]+)$`)
	return r.MatchString(snapRevision)
}

func IsHexString(s string, l int) bool {
	r := regexp.MustCompile(fmt.Sprintf("^[a-fA-F0-9]{%d}$", l))
	return r.MatchString(s)
}

//func StringToUUID(seed string) uuid.UUID {
//	// Generate UUID using NewV3 or NewV5 function
//	u := uuid.NewV3(uuid.NamespaceDNS, seed)
//	return u
//}

func MatchWithWildcards(name, pattern string) bool {
	re := regexp.MustCompile("^" + strings.Replace(pattern, "*", ".*", -1) + "$")
	return re.MatchString(name)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func ShortenRight(s string, l int) string {
	s = strings.TrimSpace(s)
	if len(s) <= l {
		return s
	}
	return s[:l-3] + "..."
}

func ReplaceLinebreaks(s string) string {
	for _, char := range []string{"\n", "\r"} {
		s = strings.ReplaceAll(s, char, " ")
	}
	return s
}

func ReformatTime(input, targetFormat string) string {
	if input == "" {
		return ""
	}

	rfc3339 := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)
	unix := regexp.MustCompile(`^\d+$`)

	switch {
	case rfc3339.MatchString(input):
		t, _ := time.Parse(time.RFC3339, input)
		return t.Format(targetFormat)
	case unix.MatchString(input):
		n, _ := strconv.ParseInt(input, 10, 64)
		t := time.Unix(n, 0)
		return t.Format(targetFormat)
	default:
		return fmt.Sprintf("%v is not a valid time format", input)
	}
}

// IsValidUuid returns true if a UUID is valid and contains dashes in the correct places according to RFC 4122.
func IsValidUuid(id string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	res := r.MatchString(id)
	if res {
		_, err := uuid.FromString(id) // does not require dashes, contrary to RFC 4122
		if err == nil {
			return true
		}
	}
	return false
}

func SplitIntoEqualChunks(s []byte, maxChunkSize int) [][]byte {
	if maxChunkSize < 1 {
		panic("maxChunkSize must be greater than zero")
	}
	if maxChunkSize >= len(s) {
		return [][]byte{s}
	}
	var chunks [][]byte = make([][]byte, 0, (len(s)-1)/maxChunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == maxChunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

func FlagToMaybeString(cmd *cobra.Command, flag string) *string {
	if cmd.Flag(flag).Changed {
		k, err := cmd.Flags().GetString(flag)
		if err != nil {
			return nil
		}
		return &k
	}
	return nil
}

func MaybeStringToMaybeTime(maybeString *string, format string) *time.Time {
	if maybeString == nil {
		return nil
	}
	parsed, err := time.Parse(format, *maybeString)
	if err != nil {
		return nil
	}
	return &parsed
}

func StrPtr(s string) *string {
	return &s
}

// Ptr returns a pointer to the given value.
// This function works with any type, not just strings.
func Ptr[T any](value T) *T {
	return &value
}
