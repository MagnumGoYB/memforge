package index

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenCreatesSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.sqlite")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, table := range []string{"schema_version", "memories", "memories_fts"} {
		var name string
		if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE name = ?`, table).Scan(&name); err != nil {
			t.Fatalf("missing %s: %v", table, err)
		}
	}
}

func TestOpenCanReplayMigrations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.sqlite")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
}

func TestOpenWaitsForTransientSchemaLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.sqlite")
	lockedDB, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer lockedDB.Close()

	tx, err := lockedDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`UPDATE schema_version SET version = ?`, CurrentSchemaVersion); err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		db, err := Open(path)
		if err == nil {
			_ = db.Close()
		}
		errCh <- err
	}()

	time.Sleep(100 * time.Millisecond)
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Open should wait for transient schema lock: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Open did not return after transient schema lock cleared")
	}
}
