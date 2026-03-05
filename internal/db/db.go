package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS env_sets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS env_vars (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    env_set_id INTEGER NOT NULL REFERENCES env_sets(id) ON DELETE CASCADE,
    key        TEXT NOT NULL,
    value      BLOB NOT NULL
);
`

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "umgebung", "umgebung.db")
}

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		return nil, err
	}
	return db, nil
}

func IsInitialized(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	db, err := Open(path)
	if err != nil {
		return false
	}
	defer db.Close()
	var salt string
	err = db.QueryRow(`SELECT value FROM meta WHERE key = 'salt'`).Scan(&salt)
	return err == nil && salt != ""
}
