package format

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 42 * time.Second, "42s"},
		{"one minute", 60 * time.Second, "1m 0s"},
		{"minutes and seconds", 45*time.Minute + 12*time.Second, "45m 12s"},
		{"exactly one hour", 60 * time.Minute, "1h 0m"},
		{"hours and minutes", 2*time.Hour + 35*time.Minute, "2h 35m"},
		{"hours minutes seconds drops seconds", 1*time.Hour + 23*time.Minute + 45*time.Second, "1h 23m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Duration(tt.duration)
			if got != tt.want {
				t.Errorf("Duration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestParseLimitDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"minutes", "25m", 25 * time.Minute, false},
		{"hours", "1h", 1 * time.Hour, false},
		{"hours and minutes", "1h30m", 1*time.Hour + 30*time.Minute, false},
		{"just number minutes", "90m", 90 * time.Minute, false},
		{"invalid", "abc", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLimit(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLimit(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseLimit(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
