package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var urgeCmd = &cobra.Command{
	Use:   "urge",
	Short: "Log a distraction urge",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := sessionSvc.LogUrge()
		if err != nil {
			return err
		}
		if result.TaskName != "" {
			fmt.Fprintf(os.Stdout, "Urge logged (during: %s).\n", result.TaskName)
		} else {
			fmt.Fprintln(os.Stdout, "Urge logged.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(urgeCmd)
}
