package db

import (
	"database/sql"
	"fmt"
	"time"
)

type EnvSet struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EnvVar struct {
	Key   string
	Value []byte // encrypted blob
}

func ListEnvSets(db *sql.DB) ([]EnvSet, error) {
	rows, err := db.Query(`SELECT id, name, created_at, updated_at FROM env_sets ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sets []EnvSet
	for rows.Next() {
		var s EnvSet
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}
	return sets, rows.Err()
}

func GetEnvSet(db *sql.DB, name string) (*EnvSet, error) {
	var s EnvSet
	err := db.QueryRow(`SELECT id, name, created_at, updated_at FROM env_sets WHERE name = ?`, name).
		Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("env set %q not found", name)
	}
	return &s, err
}

func CreateEnvSet(db *sql.DB, name string, vars []EnvVar) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO env_sets (name) VALUES (?)`, name)
	if err != nil {
		return fmt.Errorf("create env set: %w", err)
	}
	id, _ := res.LastInsertId()
	for _, v := range vars {
		if _, err := tx.Exec(`INSERT INTO env_vars (env_set_id, key, value) VALUES (?, ?, ?)`, id, v.Key, v.Value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func UpdateEnvSet(db *sql.DB, name string, vars []EnvVar) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id int64
	if err := tx.QueryRow(`SELECT id FROM env_sets WHERE name = ?`, name).Scan(&id); err != nil {
		return fmt.Errorf("env set %q not found", name)
	}
	if _, err := tx.Exec(`DELETE FROM env_vars WHERE env_set_id = ?`, id); err != nil {
		return err
	}
	for _, v := range vars {
		if _, err := tx.Exec(`INSERT INTO env_vars (env_set_id, key, value) VALUES (?, ?, ?)`, id, v.Key, v.Value); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE env_sets SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func DeleteEnvSet(db *sql.DB, name string) error {
	res, err := db.Exec(`DELETE FROM env_sets WHERE name = ?`, name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("env set %q not found", name)
	}
	return nil
}

func GetEnvVars(db *sql.DB, envSetName string) ([]EnvVar, error) {
	row := db.QueryRow(`SELECT id FROM env_sets WHERE name = ?`, envSetName)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("env set %q not found", envSetName)
	}
	rows, err := db.Query(`SELECT key, value FROM env_vars WHERE env_set_id = ? ORDER BY key`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vars []EnvVar
	for rows.Next() {
		var v EnvVar
		if err := rows.Scan(&v.Key, &v.Value); err != nil {
			return nil, err
		}
		vars = append(vars, v)
	}
	return vars, rows.Err()
}

func GetMeta(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, key).Scan(&value)
	return value, err
}

func SetMeta(db *sql.DB, key, value string) error {
	_, err := db.Exec(`INSERT INTO meta (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}
