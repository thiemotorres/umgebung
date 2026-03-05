package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/editor"
)

var importCmd = &cobra.Command{
	Use:   "import NAME FILE",
	Short: "Import a .env file as an env set",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, file := args[0], args[1]

		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		pairs, err := editor.ParseLines(string(content))
		if err != nil {
			return err
		}
		if len(pairs) == 0 {
			return fmt.Errorf("no variables found in %s", file)
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

		var vars []db.EnvVar
		for _, p := range pairs {
			enc, err := crypto.Encrypt(key, []byte(p.Value))
			if err != nil {
				return err
			}
			vars = append(vars, db.EnvVar{Key: p.Key, Value: enc})
		}

		if err := db.CreateEnvSet(conn, name, vars); err != nil {
			return err
		}
		fmt.Printf("Imported %d variables from %s into env set %q.\n", len(vars), file, name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
