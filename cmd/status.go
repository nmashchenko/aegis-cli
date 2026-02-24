package cmd

import (
	"fmt"
	"os"

	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current task status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := sessionSvc.Status()
		if err != nil {
			return err
		}
		if !result.Active {
			fmt.Fprintln(os.Stdout, "No active task.")
			return nil
		}
		fmt.Fprintf(os.Stdout, "Active task: %s\nElapsed: %s\n", result.TaskName, format.Duration(result.Elapsed))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
