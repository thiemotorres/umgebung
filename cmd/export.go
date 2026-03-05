package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/editor"
)

var exportCmd = &cobra.Command{
	Use:   "export NAME FILE",
	Short: "Export an env set to a .env file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, file := args[0], args[1]

		// Warn user
		fmt.Printf("WARNING: This will write plaintext secrets to %s\n", file)
		fmt.Print("Continue? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
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

		vars, err := db.GetEnvVars(conn, name)
		if err != nil {
			return err
		}

		var pairs []editor.EnvPair
		for _, v := range vars {
			plaintext, err := crypto.Decrypt(key, v.Value)
			if err != nil {
				return fmt.Errorf("decrypt %s: %w", v.Key, err)
			}
			pairs = append(pairs, editor.EnvPair{Key: v.Key, Value: string(plaintext)})
		}

		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		w.WriteString(editor.FormatLines(pairs))
		if err := w.Flush(); err != nil {
			return err
		}

		fmt.Printf("Exported %d variables to %s\n", len(pairs), file)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
