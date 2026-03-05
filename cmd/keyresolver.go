package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/feto/umgebung/internal/agent"
	"github.com/feto/umgebung/internal/crypto"
	"github.com/feto/umgebung/internal/db"
	"golang.org/x/term"
)

// resolveKey ensures the agent is running and returns the encryption key.
// If the agent is not yet unlocked, it prompts for the master password.
func resolveKey() ([]byte, error) {
	bin, err := os.Executable()
	if err != nil {
		return nil, err
	}
	if err := agent.EnsureRunning(bin); err != nil {
		return nil, err
	}

	key, err := agent.GetKey()
	if err == nil {
		return key, nil
	}

	// Not unlocked - prompt for password
	fmt.Print("Password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, err
	}

	conn, err := db.Open(db.DefaultPath())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	saltHex, err := db.GetMeta(conn, "salt")
	if err != nil {
		return nil, fmt.Errorf("get salt: %w", err)
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return nil, err
	}

	derivedKey := crypto.DeriveKey(string(pw), salt)

	if err := agent.Unlock(string(pw)); err != nil {
		return nil, err
	}

	return derivedKey, nil
}
