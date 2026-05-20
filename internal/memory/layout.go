package memory

import (
	"os"
	"path/filepath"
)

func EnsureLayout(memoriesDir string) error {
	if err := os.MkdirAll(memoriesDir, 0o755); err != nil {
		return err
	}
	for _, file := range DefaultFileNames() {
		path := filepath.Join(memoriesDir, file)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			return err
		}
	}
	return nil
}
