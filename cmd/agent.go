package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thiemotorres/umgebung/internal/agent"
	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
)

var agentCmd = &cobra.Command{
	Use:    "__agent",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := db.Open(db.DefaultPath())
		if err != nil {
			return err
		}
		saltHex, err := db.GetMeta(conn, "salt")
		conn.Close()
		if err != nil {
			return fmt.Errorf("get salt: %w", err)
		}
		salt, err := hex.DecodeString(saltHex)
		if err != nil {
			return err
		}

		deriveFn := func(password string) ([]byte, error) {
			return crypto.DeriveKey(password, salt), nil
		}

		return agent.RunDaemon(deriveFn)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
