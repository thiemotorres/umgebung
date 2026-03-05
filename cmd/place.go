package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/db"
)

var placeCmd = &cobra.Command{
	Use:   "place NAME",
	Short: "Write a .umgebung file in the current directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Verify the env set exists
		conn, err := db.Open(db.DefaultPath())
		if err != nil {
			return err
		}
		defer conn.Close()
		if _, err := db.GetEnvSet(conn, name); err != nil {
			return err
		}

		if err := os.WriteFile(".umgebung", []byte(name+"\n"), 0644); err != nil {
			return err
		}
		fmt.Printf("Wrote .umgebung with env set %q\n", name)
		fmt.Println("Run 'umgebung up' in this directory to activate.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(placeCmd)
}
