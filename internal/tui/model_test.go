package tui_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"

	"github.com/thiemotorres/umgebung/internal/db"
	"github.com/thiemotorres/umgebung/internal/tui"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func sendKey(m tui.Model, key string) tui.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(tui.Model)
}

func sendSpecialKey(m tui.Model, keyType tea.KeyType) tui.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated.(tui.Model)
}

func TestEnterKeyWithNoSets(t *testing.T) {
	conn := openTestDB(t)
	m := tui.New(conn, make([]byte, 32))
	// With no sets, enter should do nothing harmful
	result := sendSpecialKey(m, tea.KeyEnter)
	if result.ActivateName() != "" {
		t.Errorf("expected empty activateName with no sets, got %q", result.ActivateName())
	}
}

func TestEnterKeySetsActivateName(t *testing.T) {
	conn := openTestDB(t)
	// Create a test env set
	err := db.CreateEnvSet(conn, "myapp", []db.EnvVar{
		{Key: "FOO", Value: []byte("encrypted")},
	})
	if err != nil {
		t.Fatalf("create env set: %v", err)
	}

	m := tui.New(conn, make([]byte, 32))
	// Load sets by simulating the init message
	updatedModel, cmd := m.Update(nil)
	m = updatedModel.(tui.Model)
	// Run the loadSets command
	if cmd != nil {
		msg := cmd()
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(tui.Model)
	}

	result := sendSpecialKey(m, tea.KeyEnter)
	if result.ActivateName() != "myapp" {
		t.Errorf("expected activateName %q, got %q", "myapp", result.ActivateName())
	}
}

func TestNKeyEntersInputMode(t *testing.T) {
	conn := openTestDB(t)
	m := tui.New(conn, make([]byte, 32))
	result := sendKey(m, "n")
	if result.Mode() != "input" {
		t.Errorf("expected mode %q after pressing n, got %q", "input", result.Mode())
	}
}

func TestInputModeAccumulatesText(t *testing.T) {
	conn := openTestDB(t)
	m := tui.New(conn, make([]byte, 32))
	m = sendKey(m, "n") // enter input mode
	m = sendKey(m, "m")
	m = sendKey(m, "y")
	m = sendKey(m, "a")
	m = sendKey(m, "p")
	m = sendKey(m, "p")
	if m.InputBuf() != "myapp" {
		t.Errorf("expected inputBuf %q, got %q", "myapp", m.InputBuf())
	}
}

func TestEscCancelsInputMode(t *testing.T) {
	conn := openTestDB(t)
	m := tui.New(conn, make([]byte, 32))
	m = sendKey(m, "n") // enter input mode
	m = sendKey(m, "f")
	m = sendKey(m, "o")
	m = sendKey(m, "o")
	result := sendSpecialKey(m, tea.KeyEsc)
	if result.Mode() != "browse" {
		t.Errorf("expected mode %q after esc, got %q", "browse", result.Mode())
	}
	if result.InputBuf() != "" {
		t.Errorf("expected empty inputBuf after esc, got %q", result.InputBuf())
	}
}

func TestBackspaceInInputMode(t *testing.T) {
	conn := openTestDB(t)
	m := tui.New(conn, make([]byte, 32))
	m = sendKey(m, "n")
	m = sendKey(m, "f")
	m = sendKey(m, "o")
	m = sendKey(m, "o")
	result := sendSpecialKey(m, tea.KeyBackspace)
	if result.InputBuf() != "fo" {
		t.Errorf("expected inputBuf %q after backspace, got %q", "fo", result.InputBuf())
	}
}
