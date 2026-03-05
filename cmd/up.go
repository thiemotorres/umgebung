package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/feto/umgebung/internal/db"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [NAME]",
	Short: "Activate an env set in a sub-shell",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		} else {
			// Read from .umgebung file
			data, err := os.ReadFile(".umgebung")
			if err != nil {
				return fmt.Errorf("no NAME given and no .umgebung file in current directory")
			}
			name = strings.TrimSpace(string(data))
		}

		key, err := resolveKey()
		if err != nil {
			return err
		}

		conn, err := db.Open(db.DefaultPath())
		if err != nil {
			return err
		}
		defer conn.Close()

		return activateEnvSet(name, key, conn)
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
