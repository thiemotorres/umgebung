package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/editor"
)

var newCmd = &cobra.Command{
	Use:   "new NAME",
	Short: "Create a new env set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		key, err := resolveKey()
		if err != nil {
			return err
		}

		pairs, err := editor.Open(nil)
		if err != nil {
			return err
		}
		if len(pairs) == 0 {
			return fmt.Errorf("no variables entered")
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
		fmt.Printf("Created env set %q with %d variables.\n", name, len(vars))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
