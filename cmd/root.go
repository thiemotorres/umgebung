package cmd

import (
	"fmt"
	"os"

	"github.com/feto/umgebung/internal/db"
	"github.com/feto/umgebung/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "umgebung",
	Short: "Encrypted environment variable manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !db.IsInitialized(db.DefaultPath()) {
			fmt.Println("umgebung is not initialized. Run: umgebung init")
			return nil
		}
		return runTUI()
	},
}

func runTUI() error {
	key, err := resolveKey()
	if err != nil {
		return err
	}
	conn, err := db.Open(db.DefaultPath())
	if err != nil {
		return err
	}
	defer conn.Close()
	return tui.Run(conn, key)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
