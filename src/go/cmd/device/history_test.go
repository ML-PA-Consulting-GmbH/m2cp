package device

import (
	"m2cpcli/backend"
	"strings"
	"testing"
)

func TestHistoryOutputFormatterRtd(t *testing.T) {
	seq1, seq2 := 1, 2
	out, err := historyOutputFormatter(historyCmdOutput{
		isRtd: true,
		RtdChanges: []backend.DeviceInstallHistoryEntry{
			{DateTime: "2026-01-02T10:00:00Z", Action: "Install", AppName: "riot-app", Version: "2.0.0", Revision: 7, SequenceNumber: &seq2},
			{DateTime: "2026-01-01T10:00:00Z", Action: "Install", AppName: "riot-app", Version: "1.0.0", Revision: 3, SequenceNumber: &seq1},
			{DateTime: "2026-01-03T10:00:00Z", Action: "Install", AppName: "riot-app", Version: "1.0.0", Revision: 3},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"Sequence", "2026-01-01 10:00:00", "2.0.0", "riot-app"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "1.0.0") > strings.Index(out, "2.0.0") {
		t.Errorf("expected oldest change first:\n%s", out)
	}
}
