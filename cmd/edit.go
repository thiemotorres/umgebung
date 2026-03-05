package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/editor"
)

var editCmd = &cobra.Command{
	Use:   "edit NAME",
	Short: "Edit an existing env set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		key, err := resolveKey()
		if err != nil {
			return err
		}

		conn, err := db.Open(db.DefaultPath())
		if err != nil {
			return err
		}
		defer conn.Close()

		existing, err := db.GetEnvVars(conn, name)
		if err != nil {
			return err
		}

		// Decrypt existing values for editor pre-fill
		var initial []editor.EnvPair
		for _, v := range existing {
			plaintext, err := crypto.Decrypt(key, v.Value)
			if err != nil {
				return fmt.Errorf("decrypt %s: %w", v.Key, err)
			}
			initial = append(initial, editor.EnvPair{Key: v.Key, Value: string(plaintext)})
		}

		pairs, err := editor.Open(initial)
		if err != nil {
			return err
		}

		var vars []db.EnvVar
		for _, p := range pairs {
			enc, err := crypto.Encrypt(key, []byte(p.Value))
			if err != nil {
				return err
			}
			vars = append(vars, db.EnvVar{Key: p.Key, Value: enc})
		}

		if err := db.UpdateEnvSet(conn, name, vars); err != nil {
			return err
		}
		fmt.Printf("Updated env set %q.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
