package stats

import (
	"fmt"
	"time"

	"github.com/nmashchenko/aegis-cli/internal/db"
	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/nmashchenko/aegis-cli/internal/models"
)

type Service struct {
	db *db.DB
}

func NewService(database *db.DB) *Service {
	return &Service{db: database}
}

func PeriodToTimeRange(period string, now time.Time) (time.Time, time.Time, string, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	switch period {
	case "", "today":
		return today, tomorrow, "Today", nil
	case "week":
		return today.Add(-6 * 24 * time.Hour), tomorrow, "Last 7 Days", nil
	case "month":
		return today.Add(-29 * 24 * time.Hour), tomorrow, "Last 30 Days", nil
	case "year":
		return today.Add(-364 * 24 * time.Hour), tomorrow, "Last 365 Days", nil
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("unknown period %q: use today, week, month, or year", period)
	}
}

func (s *Service) GetFormattedStats(period string) (string, error) {
	now := time.Now()
	start, end, label, err := PeriodToTimeRange(period, now)
	if err != nil {
		return "", err
	}

	stats, err := s.db.GetStats(start, end)
	if err != nil {
		return "", fmt.Errorf("get stats: %w", err)
	}

	return FormatStats(stats, label), nil
}

func FormatStats(stats *models.Stats, label string) string {
	totalDuration := time.Duration(stats.TotalFocusSecs) * time.Second
	avgDuration := time.Duration(stats.AvgTaskSecs) * time.Second

	return fmt.Sprintf(`%s:
  Tasks Completed: %d
  Total Focus Time: %s
  Avg Task Duration: %s
  Urges Logged: %d`,
		label,
		stats.TasksCompleted,
		format.Duration(totalDuration),
		format.Duration(avgDuration),
		stats.UrgesLogged,
	)
}
