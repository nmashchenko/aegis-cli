package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the active focus task",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := sessionSvc.Pause()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Task paused: %s\n", result.TaskName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}
