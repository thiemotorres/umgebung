package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Deactivate the current env set",
	RunE: func(cmd *cobra.Command, args []string) error {
		active := os.Getenv("UMGEBUNG_ACTIVE")
		if active == "" {
			fmt.Println("No env set is currently active.")
			return nil
		}
		fmt.Printf("Active env set: %q\n", active)
		fmt.Println("Type 'exit' to leave the sub-shell and deactivate.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
