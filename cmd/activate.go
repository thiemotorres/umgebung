package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thiemotorres/umgebung/internal/crypto"
	"github.com/thiemotorres/umgebung/internal/db"
)

// activateEnvSet decrypts and injects vars from the named env set, then spawns a sub-shell.
func activateEnvSet(name string, key []byte, conn *sql.DB) error {
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
}
