package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nmashchenko/aegis-cli/internal/db"
	"github.com/nmashchenko/aegis-cli/internal/session"
	"github.com/nmashchenko/aegis-cli/internal/stats"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var (
	database   *db.DB
	sessionSvc *session.Service
	statsSvc   *stats.Service
)

var rootCmd = &cobra.Command{
	Use:   "aegis",
	Short: "Track focus tasks and distraction urges",
	Long:  "aegis-cli is a lightweight CLI tool for tracking deep work sessions and distraction impulses.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner()
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := db.DefaultPath()
		if err != nil {
			return err
		}
		database, err = db.New(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		sessionSvc = session.NewService(database)
		statsSvc = stats.NewService(database)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if database != nil {
			database.Close()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printBanner() {
	ascii := `
  ██████╗ ███████╗ ██████╗ ██╗███████╗
 ██╔══██╗██╔════╝██╔════╝ ██║██╔════╝
 ███████║█████╗  ██║  ███╗██║███████╗
 ██╔══██║██╔══╝  ██║   ██║██║╚════██║
 ██║  ██║███████╗╚██████╔╝██║███████║
 ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝╚══════╝`

	gradientColors := []string{"#5B86E5", "#7B6BE5", "#A855F7", "#9B5DE5", "#5B86E5"}
	lines := splitLines(ascii)

	fmt.Fprintln(os.Stdout)
	for i, line := range lines {
		ci := i % len(gradientColors)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(gradientColors[ci])).Bold(true)
		fmt.Fprintln(os.Stdout, style.Render(line))
	}

	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	fmt.Fprintln(os.Stdout, tagline.Render("  Shield your focus. Track your urges."))
	fmt.Fprintln(os.Stdout)

	// Weekly stats
	now := time.Now()
	start, end, _, err := stats.PeriodToTimeRange("week", now)
	if err == nil {
		st, err := database.GetStats(start, end)
		if err == nil {
			statBox := renderStatLine(st.TasksCompleted, st.UrgesLogged)
			fmt.Fprintln(os.Stdout, statBox)
			fmt.Fprintln(os.Stdout)
		}
	}

	// Commands hint
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fmt.Fprintln(os.Stdout, hint.Render("  Commands:"))

	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	fmt.Fprintf(os.Stdout, "    %s  %s\n", cmdStyle.Render("aegis start <task>"), descStyle.Render("Start a focus session"))
	fmt.Fprintf(os.Stdout, "    %s  %s\n", cmdStyle.Render("aegis stop         "), descStyle.Render("Stop current session"))
	fmt.Fprintf(os.Stdout, "    %s  %s\n", cmdStyle.Render("aegis history      "), descStyle.Render("Browse past sessions"))
	fmt.Fprintf(os.Stdout, "    %s  %s\n", cmdStyle.Render("aegis stats        "), descStyle.Render("View statistics"))
	fmt.Fprintln(os.Stdout)
}

func renderStatLine(tasks int, urges int) string {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	val := lipgloss.NewStyle().Bold(true)

	taskColor := lipgloss.Color("#22C55E")
	urgeColor := lipgloss.Color("#22C55E")
	if urges > 10 {
		urgeColor = lipgloss.Color("#F59E0B")
	}
	if urges > 25 {
		urgeColor = lipgloss.Color("#EF4444")
	}

	return fmt.Sprintf("  %s %s    %s %s",
		label.Render("This week:"),
		val.Foreground(taskColor).Render(fmt.Sprintf("%d tasks", tasks)),
		label.Render("Urges:"),
		val.Foreground(urgeColor).Render(fmt.Sprintf("%d", urges)),
	)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
