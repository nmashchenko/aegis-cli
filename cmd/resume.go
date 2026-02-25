package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume the paused focus task",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := sessionSvc.Resume()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Task resumed: %s\n", result.TaskName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
