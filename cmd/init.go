package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize umgebung database",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := db.DefaultPath()
		if db.IsInitialized(path) {
			return fmt.Errorf("already initialized at %s", path)
		}

		fmt.Print("Set master password: ")
		pw1, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return err
		}

		fmt.Print("Confirm password: ")
		pw2, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return err
		}

		if string(pw1) != string(pw2) {
			return fmt.Errorf("passwords do not match")
		}
		if len(pw1) == 0 {
			return fmt.Errorf("password cannot be empty")
		}

		salt := crypto.GenerateSalt()

		conn, err := db.Open(path)
		if err != nil {
			return err
		}
		defer conn.Close()

		if err := db.SetMeta(conn, "salt", hex.EncodeToString(salt)); err != nil {
			return err
		}

		fmt.Printf("Initialized at %s\n", path)
		fmt.Println("Run 'umgebung new <name>' to create your first env set.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
