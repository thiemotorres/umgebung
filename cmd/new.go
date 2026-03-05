package cmd

import (
	"fmt"

	"github.com/feto/umgebung/internal/crypto"
	"github.com/feto/umgebung/internal/db"
	"github.com/feto/umgebung/internal/editor"
	"github.com/spf13/cobra"
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
