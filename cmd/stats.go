package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [today|week|month|year]",
	Short: "Show focus stats",
	Long:  "Show aggregated focus stats. Defaults to today.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		period := ""
		if len(args) > 0 {
			period = args[0]
		}

		output, err := statsSvc.GetFormattedStats(period)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
