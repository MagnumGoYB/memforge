package project

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Meta struct {
	ProjectID     string    `json:"project_id"`
	ProjectRoot   string    `json:"project_root"`
	Identifier    string    `json:"identifier"`
	SchemaVersion string    `json:"schema_version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func LoadMeta(path string) (Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, err
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return Meta{}, err
	}
	return meta, nil
}

func WriteMeta(path string, meta Meta, now time.Time) (Meta, error) {
	existing, err := LoadMeta(path)
	if err == nil {
		meta.CreatedAt = existing.CreatedAt
	} else if !errors.Is(err, os.ErrNotExist) {
		return Meta{}, err
	} else {
		meta.CreatedAt = now.UTC()
	}
	meta.UpdatedAt = now.UTC()
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return Meta{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Meta{}, err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return Meta{}, err
	}
	return meta, nil
}
