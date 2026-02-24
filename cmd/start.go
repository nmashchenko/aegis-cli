package cmd

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/nmashchenko/aegis-cli/internal/tui"
	"github.com/spf13/cobra"
)

var (
	detached bool
	limit    string
)

var startCmd = &cobra.Command{
	Use:   "start [task name]",
	Short: "Start a focus task",
	Long:  "Start a timed focus task. Opens a live TUI by default. Use -d for detached mode.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName := args[0]

		var limitSeconds *int64
		var limitDuration *time.Duration
		if limit != "" {
			duration, err := format.ParseLimit(limit)
			if err != nil {
				return err
			}
			secs := int64(duration.Seconds())
			limitSeconds = &secs
			limitDuration = &duration
		}

		result, err := sessionSvc.Start(taskName, limitSeconds)
		if err != nil {
			return err
		}

		if detached {
			fmt.Fprintf(os.Stdout, "Task started: %s\n", result.TaskName)
			return nil
		}

		// Live TUI mode
		model := tui.NewModel(taskName, result.TaskID, time.Now(), limitDuration, sessionSvc)
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
}

func init() {
	startCmd.Flags().BoolVarP(&detached, "detach", "d", false, "Run in detached/background mode")
	startCmd.Flags().StringVar(&limit, "limit", "", "Time limit for the task (e.g., 25m, 1h, 1h30m)")
	rootCmd.AddCommand(startCmd)
}
