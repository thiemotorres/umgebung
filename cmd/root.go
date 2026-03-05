package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/tui"
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

	activateName, err := tui.Run(conn, key)
	if err != nil {
		return err
	}
	if activateName != "" {
		return activateEnvSet(activateName, key, conn)
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
