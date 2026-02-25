package stats

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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
	case "today", "day":
		return today, tomorrow, "Today", nil
	case "", "week":
		return today.Add(-6 * 24 * time.Hour), tomorrow, "Last 7 Days", nil
	case "month":
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)
		return monthStart, monthEnd, now.Format("January 2006"), nil
	case "year":
		yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		yearEnd := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
		return yearStart, yearEnd, fmt.Sprintf("%d", now.Year()), nil
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("unknown period %q: use day, week, month, or year", period)
	}
}

// normalizePeriod returns the canonical period name.
func normalizePeriod(period string) string {
	if period == "" {
		return "week"
	}
	if period == "today" {
		return "day"
	}
	return period
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

	p := normalizePeriod(period)

	// day and month: summary only
	if p == "day" || p == "month" {
		return formatSummary(stats, label), nil
	}

	// week and year: summary + breakdown
	dailyUrges, err := s.db.GetDailyUrgeCounts(start, end)
	if err != nil {
		return "", fmt.Errorf("get daily urge counts: %w", err)
	}

	dailyTasks, err := s.db.GetDailyTaskCounts(start, end)
	if err != nil {
		return "", fmt.Errorf("get daily task counts: %w", err)
	}

	var buckets []models.StatBucket
	if p == "week" {
		buckets = dailyBuckets(dailyTasks, dailyUrges)
	} else {
		buckets = monthlyBuckets(dailyTasks, dailyUrges)
	}

	return formatWithBreakdown(stats, label, buckets), nil
}

func dailyBuckets(tasks, urges []models.DailyUrgeCount) []models.StatBucket {
	buckets := make([]models.StatBucket, len(urges))
	taskMap := make(map[string]int, len(tasks))
	for _, t := range tasks {
		taskMap[t.Date.Format("2006-01-02")] = t.Count
	}
	for i, u := range urges {
		key := u.Date.Format("2006-01-02")
		buckets[i] = models.StatBucket{
			Label: u.Date.Format("Mon 02"),
			Tasks: taskMap[key],
			Urges: u.Count,
		}
	}
	return buckets
}

func monthlyBuckets(tasks, urges []models.DailyUrgeCount) []models.StatBucket {
	type monthKey struct {
		Year  int
		Month time.Month
	}

	var order []monthKey
	taskSums := make(map[monthKey]int)
	urgeSums := make(map[monthKey]int)

	for _, t := range tasks {
		k := monthKey{t.Date.Year(), t.Date.Month()}
		if _, exists := taskSums[k]; !exists {
			order = append(order, k)
		}
		taskSums[k] += t.Count
	}
	for _, u := range urges {
		k := monthKey{u.Date.Year(), u.Date.Month()}
		if _, exists := urgeSums[k]; !exists {
			if _, exists2 := taskSums[k]; !exists2 {
				order = append(order, k)
			}
		}
		urgeSums[k] += u.Count
	}

	buckets := make([]models.StatBucket, len(order))
	for i, k := range order {
		buckets[i] = models.StatBucket{
			Label: k.Month.String()[:3],
			Tasks: taskSums[k],
			Urges: urgeSums[k],
		}
	}
	return buckets
}

func renderStatsGrid(stats *models.Stats) string {
	totalDuration := time.Duration(stats.TotalFocusSecs) * time.Second
	avgDuration := time.Duration(stats.AvgTaskSecs) * time.Second

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	urgeValStyle := valueStyle
	if stats.UrgesLogged > 0 {
		urgeValStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444"))
	}

	type cell struct {
		label string
		value string
		style lipgloss.Style
	}

	// 2x2 grid: [row][col]
	grid := [2][2]cell{
		{
			{"Tasks:", fmt.Sprintf("%d", stats.TasksCompleted), valueStyle},
			{"Focus:", format.Duration(totalDuration), valueStyle},
		},
		{
			{"Avg:", format.Duration(avgDuration), valueStyle},
			{"Urges:", fmt.Sprintf("%d", stats.UrgesLogged), urgeValStyle},
		},
	}

	// Compute max label width per column
	colLabelW := [2]int{}
	colValW := [2]int{}
	for _, row := range grid {
		for c, cell := range row {
			if len(cell.label) > colLabelW[c] {
				colLabelW[c] = len(cell.label)
			}
			if len(cell.value) > colValW[c] {
				colValW[c] = len(cell.value)
			}
		}
	}

	var b strings.Builder
	for i, row := range grid {
		b.WriteString("  ")
		for c, cell := range row {
			labelPad := strings.Repeat(" ", colLabelW[c]-len(cell.label))
			valPad := strings.Repeat(" ", colValW[c]-len(cell.value))
			b.WriteString(fmt.Sprintf("%s%s %s%s",
				labelStyle.Render(cell.label), labelPad,
				valPad, cell.style.Render(cell.value),
			))
			if c == 0 {
				b.WriteString("    ")
			}
		}
		if i == 0 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// formatSummary renders a compact summary for day/month views.
func formatSummary(stats *models.Stats, label string) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A855F7"))

	var b strings.Builder
	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render(fmt.Sprintf("Stats · %s", label)))
	b.WriteString("\n\n")
	b.WriteString(renderStatsGrid(stats))
	b.WriteString("\n")
	return b.String()
}

// formatWithBreakdown renders summary + per-bucket table for week/year views.
func formatWithBreakdown(stats *models.Stats, label string, buckets []models.StatBucket) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A855F7"))

	var b strings.Builder
	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render(fmt.Sprintf("Stats · %s", label)))
	b.WriteString("\n\n")
	b.WriteString(renderStatsGrid(stats))

	if len(buckets) > 0 {
		b.WriteString("\n\n")
		b.WriteString(renderBreakdown(buckets))
	}

	b.WriteString("\n")
	return b.String()
}

func renderBreakdown(buckets []models.StatBucket) string {
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Column widths = max of header name and widest data value
	maxLabelLen := 0
	taskColW := len("tasks")
	urgeColW := len("urges")
	for _, b := range buckets {
		if len(b.Label) > maxLabelLen {
			maxLabelLen = len(b.Label)
		}
		if w := len(fmt.Sprintf("%d", b.Tasks)); w > taskColW {
			taskColW = w
		}
		if w := len(fmt.Sprintf("%d", b.Urges)); w > urgeColW {
			urgeColW = w
		}
	}

	var sb strings.Builder

	// Table header — pad manually to avoid ANSI width issues
	labelPad := strings.Repeat(" ", maxLabelLen)
	taskHdrPad := strings.Repeat(" ", taskColW-len("tasks"))
	urgeHdrPad := strings.Repeat(" ", urgeColW-len("urges"))
	sb.WriteString(fmt.Sprintf("  %s  %s%s  %s%s\n",
		labelPad,
		taskHdrPad, headerStyle.Render("tasks"),
		urgeHdrPad, headerStyle.Render("urges"),
	))

	for _, bucket := range buckets {
		// Color urge number: dim (0-2), amber (3-5), bold red (6+)
		var urgeStyle lipgloss.Style
		switch {
		case bucket.Urges >= 6:
			urgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
		case bucket.Urges >= 3:
			urgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		default:
			urgeStyle = dimStyle
		}

		pad := strings.Repeat(" ", maxLabelLen-len(bucket.Label))
		taskStr := fmt.Sprintf("%*d", taskColW, bucket.Tasks)
		urgeStr := fmt.Sprintf("%*d", urgeColW, bucket.Urges)

		sb.WriteString(fmt.Sprintf("  %s%s  %s  %s\n",
			bucket.Label, pad,
			dimStyle.Render(taskStr),
			urgeStyle.Render(urgeStr),
		))
	}

	return sb.String()
}
