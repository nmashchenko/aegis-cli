package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/nmashchenko/aegis-cli/internal/updater"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current version",
	RunE: func(cmd *cobra.Command, args []string) error {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
		greenStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22C55E"))

		fmt.Fprintf(os.Stdout, "%s %s\n",
			labelStyle.Render("aegis"),
			valueStyle.Render(Version),
		)

		latest, err := updater.GetLatestVersion()
		if err != nil {
			return nil // silently skip if offline
		}

		if updater.IsNewer(Version, latest) {
			fmt.Fprintf(os.Stdout, "%s %s\n",
				labelStyle.Render("Update available:"),
				greenStyle.Render(latest),
			)
			fmt.Fprintf(os.Stdout, "%s\n",
				labelStyle.Render("Run \"aegis update\" to install"),
			)
		} else {
			fmt.Fprintf(os.Stdout, "%s\n",
				greenStyle.Render("Up to date"),
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
