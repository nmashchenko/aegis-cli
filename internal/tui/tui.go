package tui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/nmashchenko/aegis-cli/internal/session"
)

// urgeMood represents the mood derived from urge count.
type urgeMood int

const (
	moodCalm    urgeMood = iota // 0-2 urges
	moodAnnoyed                 // 3-4 urges
	moodAngry                   // 5-9 urges
	moodChaos                   // 10+ urges
)

func deriveMood(urgeCount int) urgeMood {
	switch {
	case urgeCount >= 10:
		return moodChaos
	case urgeCount >= 5:
		return moodAngry
	case urgeCount >= 3:
		return moodAnnoyed
	default:
		return moodCalm
	}
}

func moodEmoji(m urgeMood) string {
	switch m {
	case moodAnnoyed:
		return " 😤"
	case moodAngry:
		return " 🔥"
	case moodChaos:
		return " 💀"
	default:
		return ""
	}
}

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

	spinner    spinner.Model
	progress   progress.Model
	frameCount int64
	width      int
	height     int
}

const (
	cardWidth   = 46
	barWidth    = 38
	tickRate    = 100 * time.Millisecond
	breathCycle = 30 // frames per full breath cycle (~3s at 10 FPS)
	gradientLen = 900 // frames for full gradient cycle (~90s at 10 FPS)
)

func NewModel(taskName string, taskID int64, startedAt time.Time, limit *time.Duration, svc *session.Service) Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{
			"█▓▒░░░░",
			"░█▓▒░░░",
			"░░█▓▒░░",
			"░░░█▓▒░",
			"░░░░█▓▒",
			"░░░░░█▓",
			"░░░░█▓▒",
			"░░░█▓▒░",
			"░░█▓▒░░",
			"░█▓▒░░░",
		},
		FPS: time.Second / 12,
	}

	p := progress.New(
		progress.WithSolidFill("#22C55E"),
		progress.WithWidth(barWidth),
		progress.WithoutPercentage(),
	)

	return Model{
		taskName:   taskName,
		taskID:     taskID,
		startedAt:  startedAt,
		limit:      limit,
		sessionSvc: svc,
		spinner:    s,
		progress:   p,
	}
}

func tickEvery() tea.Cmd {
	return tea.Tick(tickRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickEvery(), m.spinner.Tick)
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
		m.frameCount++

		var cmds []tea.Cmd
		cmds = append(cmds, tickEvery())

		if m.limit != nil {
			elapsed := time.Since(m.startedAt)
			ratio := float64(elapsed) / float64(*m.limit)
			if ratio > 1 {
				ratio = 1
			}
			cmds = append(cmds, m.progress.SetPercent(ratio))
		}

		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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

	mood := deriveMood(m.urgeCount)
	elapsed := time.Since(m.startedAt)

	// --- Compute colors ---
	borderColor := m.gradientBorderColor(mood)
	timerColor := m.breathingColor(borderColor, mood)

	// --- Build card content ---
	var lines []string

	// Empty top padding
	lines = append(lines, "")

	// Spinner + task name
	spinnerStr := m.styledSpinner(mood, borderColor)
	taskLine := spinnerStr + " " + lipgloss.NewStyle().Bold(true).Render(m.taskName)
	lines = append(lines, taskLine)

	lines = append(lines, "")

	// Timer line
	timerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(timerColor.Hex())).Bold(true)
	if m.limit != nil {
		limitDur := *m.limit
		timerStr := timerStyle.Render(fmt.Sprintf("%s / %s", formatTimer(elapsed), formatTimer(limitDur)))
		if elapsed > limitDur {
			overtime := elapsed - limitDur
			overtimeStr := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4444")).
				Bold(true).
				Render(fmt.Sprintf("  OVERTIME +%s", formatTimer(overtime)))
			timerStr += overtimeStr
		}
		lines = append(lines, timerStr)
	} else {
		lines = append(lines, timerStyle.Render(formatTimer(elapsed)))
	}

	// Urge line
	urgeText := fmt.Sprintf("Urges: %d%s", m.urgeCount, moodEmoji(mood))
	urgeStyle := m.urgeStyle(mood)
	lines = append(lines, urgeStyle.Render(urgeText))

	lines = append(lines, "")

	// Progress bar (only with limit)
	if m.limit != nil {
		m.applyProgressColors(mood)
		lines = append(lines, m.progress.View())
		lines = append(lines, "")
	}

	// Help footer
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	lines = append(lines, helpStyle.Render("u: log urge  |  q: stop session"))

	lines = append(lines, "")

	content := strings.Join(lines, "\n")

	// --- Card border ---
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.cardBorderHex(mood, borderColor))).
		Padding(0, 2).
		Width(cardWidth)

	card := borderStyle.Render(content)

	return "\n" + card + "\n"
}

// --- Color helpers ---

// gradientBorderColor cycles blue → purple → teal over ~90s.
func (m Model) gradientBorderColor(mood urgeMood) colorful.Color {
	keyColors := []colorful.Color{
		mustParseHex("#5B86E5"), // blue
		mustParseHex("#A855F7"), // purple
		mustParseHex("#2DD4BF"), // teal
	}

	t := float64(m.frameCount%int64(gradientLen)) / float64(gradientLen)
	segments := len(keyColors)
	scaledT := t * float64(segments)
	idx := int(scaledT) % segments
	frac := scaledT - float64(int(scaledT))
	next := (idx + 1) % segments

	return keyColors[idx].BlendLab(keyColors[next], frac)
}

// breathingColor modulates lightness with a sine wave.
func (m Model) breathingColor(base colorful.Color, mood urgeMood) colorful.Color {
	phase := float64(m.frameCount) * 2 * math.Pi / float64(breathCycle)
	delta := math.Sin(phase) * 0.08

	h, s, l := base.Hsl()
	l = clamp01(l + delta)

	return colorful.Hsl(h, s, l)
}

// styledSpinner returns spinner text colored by mood.
func (m Model) styledSpinner(mood urgeMood, borderColor colorful.Color) string {
	var hex string
	switch mood {
	case moodAnnoyed:
		hex = "#F59E0B"
	case moodAngry:
		hex = "#EF4444"
	case moodChaos:
		hex = "#FF0000"
	default:
		hex = borderColor.Hex()
	}
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
	return m.spinner.View()
}

// urgeStyle returns a style for the urge count based on mood.
func (m Model) urgeStyle(mood urgeMood) lipgloss.Style {
	switch mood {
	case moodAnnoyed:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	case moodAngry:
		// Pulsing red: modulate brightness
		phase := float64(m.frameCount) * 2 * math.Pi / float64(breathCycle)
		l := 0.5 + math.Sin(phase)*0.15
		c := colorful.Hsl(0, 0.85, clamp01(l))
		return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex()))
	case moodChaos:
		// Wobble/jitter: randomly shift foreground between bright red shades
		r := rand.Float64()*0.3 + 0.7
		c := colorful.Color{R: r, G: 0, B: 0}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
	}
}

// cardBorderHex returns the border color hex string, overridden by mood.
func (m Model) cardBorderHex(mood urgeMood, gradientColor colorful.Color) string {
	switch mood {
	case moodAnnoyed:
		return "#F59E0B"
	case moodAngry:
		// Breathing red border
		phase := float64(m.frameCount) * 2 * math.Pi / float64(breathCycle)
		l := 0.45 + math.Sin(phase)*0.1
		c := colorful.Hsl(0, 0.85, clamp01(l))
		return c.Hex()
	case moodChaos:
		// Flickering border: random bright red variation each frame
		r := rand.Float64()*0.4 + 0.6
		g := rand.Float64() * 0.1
		c := colorful.Color{R: r, G: g, B: 0}
		return c.Hex()
	default:
		return gradientColor.Hex()
	}
}

// applyProgressColors sets the progress bar gradient based on mood.
func (m *Model) applyProgressColors(mood urgeMood) {
	switch mood {
	case moodAnnoyed:
		m.progress.FullColor = "#F59E0B"
		m.progress.EmptyColor = "#78716C"
	case moodAngry:
		m.progress.FullColor = "#EF4444"
		m.progress.EmptyColor = "#78716C"
	case moodChaos:
		m.progress.FullColor = "#FF0000"
		m.progress.EmptyColor = "#450A0A"
	default:
		m.progress.FullColor = "#22C55E"
		m.progress.EmptyColor = "#1C1917"
	}
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

func mustParseHex(hex string) colorful.Color {
	c, err := colorful.Hex(hex)
	if err != nil {
		panic(fmt.Sprintf("invalid hex color: %s", hex))
	}
	return c
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
