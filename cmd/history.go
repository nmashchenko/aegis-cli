package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/nmashchenko/aegis-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	cardWidth = 44

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1).
			MarginBottom(1)

	cardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(cardWidth).
			MarginBottom(0)

	cardBorderDim = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			Width(cardWidth).
			MarginBottom(0)

	taskNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	durationValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("42"))

	urgeValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	urgeValueZeroStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("42"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show recent completed tasks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetRecentTasks(5)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Fprintln(os.Stdout)
			fmt.Fprintln(os.Stdout, emptyStyle.Render("No completed tasks yet. Start one with: aegis start \"task name\""))
			fmt.Fprintln(os.Stdout)
			return nil
		}

		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, headerStyle.Render("  Recent Tasks  "))
		fmt.Fprintln(os.Stdout)

		for i, t := range tasks {
			card := renderCard(t, i)
			fmt.Fprintln(os.Stdout, card)

			// Small spacing between cards
			if i < len(tasks)-1 {
				time.Sleep(80 * time.Millisecond)
			}
		}

		fmt.Fprintln(os.Stdout, footerStyle.Render(fmt.Sprintf("  Showing %d most recent tasks", len(tasks))))
		fmt.Fprintln(os.Stdout)
		return nil
	},
}

func renderCard(t models.TaskHistory, index int) string {
	duration := time.Duration(t.DurationSeconds) * time.Second

	// Task name with number
	name := taskNameStyle.Render(fmt.Sprintf("#%d  %s", index+1, t.Name))

	// Separator line
	sep := labelStyle.Render(strings.Repeat("─", cardWidth-6))

	// Duration line
	durationLine := labelStyle.Render("Duration  ") + durationValueStyle.Render(format.Duration(duration))

	// Urges line with color based on count
	urgeStr := fmt.Sprintf("%d", t.UrgeCount)
	if t.UrgeCount == 0 {
		urgeStr = urgeValueZeroStyle.Render(urgeStr)
	} else {
		urgeStr = urgeValueStyle.Render(urgeStr)
	}
	urgeLine := labelStyle.Render("Urges     ") + urgeStr

	// Time line with relative time
	timeStr := formatRelativeTime(t.StartedAt)
	timeLine := labelStyle.Render("Started   ") + timeStyle.Render(timeStr)

	content := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", name, sep, durationLine, urgeLine, timeLine)

	// First card gets highlight border, rest get dim
	if index == 0 {
		return cardBorder.Render(content)
	}
	return cardBorderDim.Render(content)
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	local := t.Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)

	timeStr := local.Format("15:04")

	if local.After(today) {
		return fmt.Sprintf("Today %s", timeStr)
	}
	if local.After(yesterday) {
		return fmt.Sprintf("Yesterday %s", timeStr)
	}
	return local.Format("Jan 2 15:04")
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
