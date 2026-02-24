package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var forceReset bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear all tasks and urges, start fresh",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceReset {
			fmt.Fprint(os.Stderr, "This will permanently delete all tasks and urges. Are you sure? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Fprintln(os.Stdout, "Cancelled.")
				return nil
			}
		}

		if err := database.ResetAll(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "All data cleared. Fresh start!")
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&forceReset, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(resetCmd)
}
