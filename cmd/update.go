package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/nmashchenko/aegis-cli/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update aegis to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		greenStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22C55E"))

		fmt.Fprintf(os.Stdout, "%s\n", labelStyle.Render("Checking for updates..."))

		latest, err := updater.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !updater.IsNewer(Version, latest) {
			fmt.Fprintf(os.Stdout, "%s %s\n",
				greenStyle.Render("Already up to date:"),
				greenStyle.Render(Version),
			)
			return nil
		}

		fmt.Fprintf(os.Stdout, "%s %s → %s\n",
			labelStyle.Render("Updating:"),
			labelStyle.Render(Version),
			greenStyle.Render(latest),
		)

		if err := updater.Update(latest); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Fprintf(os.Stdout, "%s\n", greenStyle.Render("Updated successfully!"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
