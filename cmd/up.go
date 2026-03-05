package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/feto/umgebung/internal/crypto"
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

		vars, err := db.GetEnvVars(conn, name)
		if err != nil {
			return err
		}

		// Build env: inherit current env, add decrypted vars
		env := os.Environ()
		for _, v := range vars {
			plaintext, err := crypto.Decrypt(key, v.Value)
			if err != nil {
				return fmt.Errorf("decrypt %s: %w", v.Key, err)
			}
			env = append(env, fmt.Sprintf("%s=%s", v.Key, string(plaintext)))
		}
		env = append(env, fmt.Sprintf("UMGEBUNG_ACTIVE=%s", name))

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}

		// Prepend (name) to prompt
		for i, e := range env {
			if strings.HasPrefix(e, "PS1=") {
				env[i] = fmt.Sprintf("PS1=(%s) %s", name, e[4:])
				break
			}
		}

		fmt.Printf("Activating env set %q. Type 'exit' to deactivate.\n", name)

		c := exec.Command(shell)
		c.Env = env
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
