package format

import (
	"fmt"
	"time"
)

// Duration formats a time.Duration into a human-readable string.
// Under 1 minute: "42s"
// Under 1 hour: "45m 12s"
// 1 hour+: "2h 35m"
func Duration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	if totalSeconds < 0 {
		totalSeconds = 0
	}

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// ParseLimit parses a limit string like "25m", "1h", "1h30m" into a time.Duration.
func ParseLimit(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty limit string")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid limit format %q: use formats like 25m, 1h, 1h30m", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("limit must be positive")
	}
	return d, nil
}
