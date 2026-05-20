package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const openRetryTimeout = 5 * time.Second

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := retryBusy(openRetryTimeout, func() error { return configure(db) }); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := retryBusy(openRetryTimeout, func() error { return migrate(db) }); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func configure(db *sql.DB) error {
	for _, pragma := range []string{
		`PRAGMA journal_mode=WAL;`,
		`PRAGMA synchronous=NORMAL;`,
		`PRAGMA busy_timeout=5000;`,
	} {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("apply %s: %w", pragma, err)
		}
	}
	return nil
}

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, stmt := range schemaStatements {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	var existing string
	err = tx.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&existing)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existing != "" && existing != CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema version %q", existing)
	}
	if existing == "" {
		if _, err := tx.Exec(`INSERT INTO schema_version(version) VALUES (?);`, CurrentSchemaVersion); err != nil {
			return err
		}
	} else if _, err := tx.Exec(`UPDATE schema_version SET version = ?;`, CurrentSchemaVersion); err != nil {
		return err
	}
	return tx.Commit()
}

func retryBusy(timeout time.Duration, fn func() error) error {
	deadline := time.Now().Add(timeout)
	var err error
	for {
		err = fn()
		if err == nil || !isBusyError(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func isBusyError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "sqlite_busy") || strings.Contains(message, "database is locked")
}
