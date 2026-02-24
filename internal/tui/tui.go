package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/nmashchenko/aegis-cli/internal/session"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	overtimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	barDoneStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	barTodoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	barOverStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type tickMsg time.Time

type Model struct {
	taskName   string
	taskID     int64
	startedAt  time.Time
	limit      *time.Duration
	urgeCount  int
	sessionSvc *session.Service
	quitting   bool
	stopResult *session.StopResult
}

func NewModel(taskName string, taskID int64, startedAt time.Time, limit *time.Duration, svc *session.Service) Model {
	return Model{
		taskName:   taskName,
		taskID:     taskID,
		startedAt:  startedAt,
		limit:      limit,
		sessionSvc: svc,
	}
}

func tickEvery() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tickEvery()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			result, err := m.sessionSvc.Stop()
			if err == nil {
				m.stopResult = result
			}
			m.quitting = true
			return m, tea.Quit
		case "u":
			_, err := m.sessionSvc.LogUrge()
			if err == nil {
				m.urgeCount++
			}
			return m, nil
		}
	case tickMsg:
		return m, tickEvery()
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting && m.stopResult != nil {
		return fmt.Sprintf("\nTask stopped: %s\nDuration: %s\n",
			m.stopResult.TaskName,
			format.Duration(m.stopResult.Duration),
		)
	}

	elapsed := time.Since(m.startedAt)
	s := "\n"

	// Title line with timer
	if m.limit != nil {
		limitDur := *m.limit
		if elapsed > limitDur {
			overtime := elapsed - limitDur
			s += titleStyle.Render(fmt.Sprintf("⏳ %s  %s / %s",
				m.taskName,
				formatTimer(elapsed),
				formatTimer(limitDur),
			))
			s += "  " + overtimeStyle.Render(fmt.Sprintf("⚡ +%s overtime", formatTimer(overtime)))
		} else {
			s += titleStyle.Render(fmt.Sprintf("⏳ %s  %s / %s",
				m.taskName,
				formatTimer(elapsed),
				formatTimer(limitDur),
			))
		}
	} else {
		s += titleStyle.Render(fmt.Sprintf("⏳ %s  %s", m.taskName, formatTimer(elapsed)))
	}
	s += "\n"

	// Urge count
	s += subtitleStyle.Render(fmt.Sprintf("  Urges: %d", m.urgeCount))
	s += "\n"

	// Progress bar (only with limit)
	if m.limit != nil {
		s += "\n"
		s += "  " + renderProgressBar(elapsed, *m.limit, 32)
		s += "\n"
	}

	// Help
	s += "\n"
	s += helpStyle.Render("  Press 'u' to log urge | 'q' to stop")
	s += "\n"

	return s
}

func formatTimer(d time.Duration) string {
	total := int(d.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%02d:%02d", m, sec)
}

func renderProgressBar(elapsed, limit time.Duration, width int) string {
	ratio := float64(elapsed) / float64(limit)
	if ratio > 1 {
		bar := ""
		for i := 0; i < width; i++ {
			bar += "█"
		}
		return barOverStyle.Render(bar) + " " + overtimeStyle.Render("OVERTIME")
	}

	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	filler := ""
	for i := 0; i < empty; i++ {
		filler += "▒"
	}

	return barDoneStyle.Render(bar) + barTodoStyle.Render(filler)
}
