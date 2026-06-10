package tools

import (
	"math"
	"strconv"
	"strings"
	"time"
)

func TimeNowF64() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}

func TimeNowDec() string {
	return TimeF64ToDec(TimeNowF64())
}

func TimeF64ToDec(t float64) string {
	s := strconv.FormatFloat(t, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func TimeToDec(t time.Time) string {
	return TimeF64ToDec(float64(t.UnixNano()) / float64(time.Second))
}

func TimeDecToTime(t string) time.Time {
	t64, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return time.Unix(0, 0).UTC()
	}
	return TimeF64ToTime(t64)

}

func TimeF64ToTime(t float64) time.Time {
	seconds := int64(t)
	nanoseconds := int64(math.Round((t - float64(seconds)) * 1e9))
	return time.Unix(seconds, nanoseconds).UTC()
}

func TimeF64ToISO8601(t float64) string {
	sec, dec := int64(t), t-float64(int64(t))
	nsec := int64(dec * 1e9)
	tUtc := time.Unix(sec, nsec).UTC()
	return tUtc.Format(time.RFC3339Nano)
}

func TimeToISO8601UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func TimeParse(value string) (time.Time, error) {
	var t time.Time
	var err error
	for _, layout := range timeLayouts {
		t, err = time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

var timeLayouts = []string{
	"2006-01-02T15:04:05.000000000Z",
	"2006-01-02T15:04:05.00000000Z",
	"2006-01-02T15:04:05.0000000Z",
	"2006-01-02T15:04:05.000000Z",
	"2006-01-02T15:04:05.00000Z",
	"2006-01-02T15:04:05.0000Z",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04Z",
	"2006-01-02",
}

func TimesSimilar(a, b time.Time, tolerance time.Duration) bool {
	delta := a.UnixNano() - b.UnixNano()
	if delta < 0 {
		delta = -delta
	}
	return time.Duration(delta) <= tolerance
}

// TimeSystemRunning returns the duration the system has been running - this is a good alternative to system time, as it's not affected by clock time changes.
func TimeSystemRunning() time.Duration {
	return timeSystemRunningOS()
}
