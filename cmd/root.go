package cmd

import (
	"fmt"
	"os"

	"github.com/nmashchenko/aegis-cli/internal/db"
	"github.com/nmashchenko/aegis-cli/internal/session"
	"github.com/nmashchenko/aegis-cli/internal/stats"
	"github.com/spf13/cobra"
)

var (
	database   *db.DB
	sessionSvc *session.Service
	statsSvc   *stats.Service
)

var rootCmd = &cobra.Command{
	Use:   "aegis",
	Short: "Track focus tasks and distraction urges",
	Long:  "aegis-cli is a lightweight CLI tool for tracking deep work sessions and distraction impulses.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := db.DefaultPath()
		if err != nil {
			return err
		}
		database, err = db.New(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		sessionSvc = session.NewService(database)
		statsSvc = stats.NewService(database)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if database != nil {
			database.Close()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
