package messages

import (
	"fmt"
	"m2cp/tools"
	"time"
)

type trace struct {
	A string `json:"A"`
	T string `json:"T"`
	t time.Time
}

func newTrace(relay string) trace {
	now := time.Now()
	return trace{
		A: relay,
		T: tools.TimeToISO8601UTC(now),
		t: now,
	}
}

func (o *trace) GetRelay() string {
	return o.A
}

func (o *trace) GetTime() (time.Time, error) {
	t, err := tools.TimeParse(o.T)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %s", err)
	}
	return t, nil
}

func (o *trace) GetTimeString() string {
	return o.T
}
