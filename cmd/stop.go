package cmd

import (
	"fmt"
	"os"

	"github.com/nmashchenko/aegis-cli/internal/format"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the active focus task",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := sessionSvc.Stop()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Task stopped: %s\nDuration: %s\n", result.TaskName, format.Duration(result.Duration))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
