package stats

import (
	"testing"
	"time"
)

func TestPeriodToTimeRange(t *testing.T) {
	now := time.Date(2026, 2, 23, 15, 30, 0, 0, time.Local)

	tests := []struct {
		period    string
		wantLabel string
		wantErr   bool
	}{
		{"today", "Today", false},
		{"week", "Last 7 Days", false},
		{"month", "Last 30 Days", false},
		{"year", "Last 365 Days", false},
		{"", "Today", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			start, end, label, err := PeriodToTimeRange(tt.period, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("PeriodToTimeRange(%q) error = %v, wantErr %v", tt.period, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if label != tt.wantLabel {
				t.Errorf("label = %q, want %q", label, tt.wantLabel)
			}
			if !start.Before(end) {
				t.Errorf("start (%v) should be before end (%v)", start, end)
			}
		})
	}
}

func TestPeriodToday(t *testing.T) {
	now := time.Date(2026, 2, 23, 15, 30, 0, 0, time.Local)
	start, end, _, err := PeriodToTimeRange("today", now)
	if err != nil {
		t.Fatal(err)
	}
	expectedStart := time.Date(2026, 2, 23, 0, 0, 0, 0, time.Local)
	expectedEnd := time.Date(2026, 2, 24, 0, 0, 0, 0, time.Local)
	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
}
