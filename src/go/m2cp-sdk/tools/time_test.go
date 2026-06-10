package tools

import (
	"strings"
	"testing"
	"time"
)

func TestTimeNowF64(t *testing.T) {
	now := time.Now()
	timeNowF64 := TimeNowF64()

	// Check if the time returned is close enough to the actual time
	if !TimesSimilar(time.Now(), TimeF64ToTime(timeNowF64), time.Millisecond) {
		t.Errorf("TimeNowF64 is not close enough to the actual time. TimeNowF64: %v, Actual: %v", timeNowF64, now)
	}
}

func TestTimeNowDec(t *testing.T) {
	now := time.Now()
	timeNowDec := TimeNowDec()

	// Check if the time returned is close enough to the actual time
	if !TimesSimilar(now, TimeDecToTime(timeNowDec), time.Millisecond) {
		t.Errorf("TimeNowDec is not close enough to the actual time. TimeNowDec: %v, Actual: %v", timeNowDec, now)
	}
}

func TestTimeToDecAndBack(t *testing.T) {
	now := time.Now()
	timeDec := TimeToDec(now)
	timeBack := TimeDecToTime(timeDec)

	// Check if the time is the same after converting back and forth
	if !TimesSimilar(now, timeBack, time.Millisecond) {
		t.Errorf("TimeToDec and TimeDecToTime conversion is not accurate. Original: %v, Converted: %v", now, timeBack)
	}
}

func TestTimeF64ToISO8601(t *testing.T) {
	now := time.Now()
	timeF64 := float64(now.UnixNano()) / float64(time.Second)
	iso8601 := TimeF64ToISO8601(timeF64)
	parsedTime, err := time.Parse(time.RFC3339Nano, iso8601)

	if err != nil {
		t.Errorf("Failed to parse the ISO8601 string: %v", iso8601)
	}

	if !TimesSimilar(now, parsedTime, time.Millisecond) {
		t.Errorf("TimeF64ToISO8601 is not accurate. Original: %v, ISO8601: %v", now, iso8601)
	}
}

func TestTimeToISO8601UTC(t *testing.T) {
	now := time.Now()
	iso8601 := TimeToISO8601UTC(now)
	parsedTime, err := time.Parse(time.RFC3339Nano, iso8601)

	if err != nil {
		t.Errorf("Failed to parse the ISO8601 string: %v", iso8601)
	}

	if !TimesSimilar(now.UTC(), parsedTime, time.Millisecond) {
		t.Errorf("TimeToISO8601UTC is not accurate. Original: %v, ISO8601: %v", now, iso8601)
	}
}

func TestTimeParse(t *testing.T) {
	now := time.Now()
	iso8601 := now.UTC().Format(time.RFC3339Nano)
	parsedTime, err := TimeParse(iso8601)

	if err != nil {
		t.Errorf("Failed to parse the ISO8601 string: %v", iso8601)
	}

	if !TimesSimilar(now.UTC(), parsedTime, time.Millisecond) {
		t.Errorf("TimeParse is not accurate. Original: %v, Parsed: %v", now, parsedTime)
	}
}

func TestTimesSimilar(t *testing.T) {
	now := time.Now()
	oneMillisecondApart := now.Add(time.Millisecond)
	oneSecondApart := now.Add(time.Second)
	if !TimesSimilar(now, oneMillisecondApart, time.Millisecond) {
		t.Errorf("TimesSimilar should return true for times within the tolerance. Time1: %v, Time2: %v", now, oneMillisecondApart)
	}

	if TimesSimilar(now, oneSecondApart, time.Millisecond) {
		t.Errorf("TimesSimilar should return false for times outside the tolerance. Time1: %v, Time2: %v", now, oneSecondApart)
	}
}

func TestTimeF64ToDec(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "Test with digits after the decimal point",
			input:    1627408512.123456,
			expected: "1627408512.123456",
		},
		{
			name:     "Test without digits after the decimal point",
			input:    1627408512.0,
			expected: "1627408512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeF64ToDec(tt.input)
			if result != tt.expected {
				t.Errorf("TimeF64ToDec(%v) = %v, want %v", tt.input, result, tt.expected)
			}

			if strings.HasSuffix(result, "0") {
				t.Errorf("TimeF64ToDec(%v) has trailing zeros: %v", tt.input, result)
			}
		})
	}
}
