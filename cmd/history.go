package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/nmashchenko/aegis-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	hdrStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("62")).Padding(0, 1)
	selStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	durStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	urgStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	urgZero   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	tmeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	delStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	ftStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type historyModel struct {
	tasks      []models.TaskHistory
	cursor     int
	deleted    string
	confirming bool
}

func (m historyModel) Init() tea.Cmd { return nil }

func (m historyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()

		// If waiting for delete confirmation
		if m.confirming {
			if s == "d" {
				task := m.tasks[m.cursor]
				if err := database.DeleteTask(task.ID); err == nil {
					m.deleted = fmt.Sprintf("Deleted \"%s\"", task.Name)
					m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
					if m.cursor >= len(m.tasks) && m.cursor > 0 {
						m.cursor--
					}
				}
			}
			m.confirming = false
			return m, nil
		}

		switch s {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.deleted = ""
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
			m.deleted = ""
		case "d":
			if len(m.tasks) > 0 {
				m.confirming = true
				m.deleted = ""
			}
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m historyModel) View() string {
	var b strings.Builder

	b.WriteString("\n  " + hdrStyle.Render(" Recent Tasks ") + "\n\n")

	if len(m.tasks) == 0 {
		if m.deleted != "" {
			b.WriteString("  " + delStyle.Render(m.deleted) + "\n\n")
		}
		b.WriteString(dimStyle.Render("  No completed tasks yet. Start one with: aegis start \"task name\"") + "\n\n")
		b.WriteString(ftStyle.Render("  q: quit") + "\n")
		return b.String()
	}

	for i, t := range m.tasks {
		cursor := "   "
		if i == m.cursor {
			cursor = selStyle.Render(" ▸ ")
		}

		dur := format.Duration(time.Duration(t.DurationSeconds) * time.Second)
		urges := fmt.Sprintf("%d", t.UrgeCount)
		ts := histRelTime(t.StartedAt)

		var nStyle lipgloss.Style
		if i == m.cursor {
			nStyle = selStyle
		} else {
			nStyle = nameStyle
		}

		// Truncate long names
		name := t.Name
		if len(name) > 24 {
			name = name[:21] + "..."
		}

		var uStyle lipgloss.Style
		if t.UrgeCount == 0 {
			uStyle = urgZero
		} else {
			uStyle = urgStyle
		}

		line := fmt.Sprintf("%s %s  %s  %s  %s",
			cursor,
			nStyle.Render(fmt.Sprintf("%-24s", name)),
			durStyle.Render(fmt.Sprintf("%8s", dur)),
			uStyle.Render(fmt.Sprintf("%2s urges", urges)),
			tmeStyle.Render(ts),
		)
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	if m.confirming {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)
		b.WriteString("  " + warnStyle.Render(fmt.Sprintf("Delete \"%s\"? Press d to confirm, any other key to cancel", m.tasks[m.cursor].Name)) + "\n")
	} else if m.deleted != "" {
		b.WriteString("  " + delStyle.Render(m.deleted) + "\n")
	}
	b.WriteString(ftStyle.Render("  ↑/↓: navigate  |  d: delete  |  q: quit") + "\n")

	return b.String()
}

func histRelTime(t time.Time) string {
	now := time.Now()
	local := t.Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)

	ts := local.Format("15:04")
	if local.After(today) {
		return "Today " + ts
	}
	if local.After(yesterday) {
		return "Yesterday " + ts
	}
	return local.Format("Jan 2 15:04")
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show recent completed tasks",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks, err := database.GetRecentTasks(20)
		if err != nil {
			return err
		}

		m := historyModel{tasks: tasks}
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("history TUI error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
