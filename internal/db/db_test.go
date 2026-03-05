package db_test

import (
	"path/filepath"
	"testing"

	"github.com/thiemotorres/umgebung/internal/db"
)

func TestOpenCreatesSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	_, err = conn.Exec(`INSERT INTO env_sets (name) VALUES ('test')`)
	if err != nil {
		t.Fatalf("insert env_set: %v", err)
	}
}

func TestIsInitialized(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	if db.IsInitialized(path) {
		t.Fatal("should not be initialized before Open")
	}

	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if db.IsInitialized(path) {
		t.Fatal("should not be initialized without salt")
	}
	conn.Exec(`INSERT INTO meta (key, value) VALUES ('salt', 'abc123')`)
	conn.Close()

	if !db.IsInitialized(path) {
		t.Fatal("should be initialized after salt set")
	}
}

func TestEnvSetCRUD(t *testing.T) {
	dir := t.TempDir()
	conn, _ := db.Open(filepath.Join(dir, "test.db"))
	defer conn.Close()

	vars := []db.EnvVar{
		{Key: "FOO", Value: []byte("encrypted-foo")},
		{Key: "BAR", Value: []byte("encrypted-bar")},
	}

	if err := db.CreateEnvSet(conn, "myapp", vars); err != nil {
		t.Fatalf("CreateEnvSet: %v", err)
	}

	sets, err := db.ListEnvSets(conn)
	if err != nil || len(sets) != 1 || sets[0].Name != "myapp" {
		t.Fatalf("ListEnvSets: %v %v", sets, err)
	}

	got, err := db.GetEnvVars(conn, "myapp")
	if err != nil || len(got) != 2 {
		t.Fatalf("GetEnvVars: %v %v", got, err)
	}

	newVars := []db.EnvVar{{Key: "BAZ", Value: []byte("encrypted-baz")}}
	if err := db.UpdateEnvSet(conn, "myapp", newVars); err != nil {
		t.Fatalf("UpdateEnvSet: %v", err)
	}
	got, _ = db.GetEnvVars(conn, "myapp")
	if len(got) != 1 || got[0].Key != "BAZ" {
		t.Fatalf("expected updated vars, got %v", got)
	}

	if err := db.DeleteEnvSet(conn, "myapp"); err != nil {
		t.Fatalf("DeleteEnvSet: %v", err)
	}
	sets, _ = db.ListEnvSets(conn)
	if len(sets) != 0 {
		t.Fatal("expected empty list after delete")
	}
}
