package memory

import (
	"fmt"
	"os"
	"path/filepath"
)

func LoadRecords(memoriesDir string, projectID string) ([]Record, error) {
	records := make([]Record, 0)
	for _, kind := range ValidKinds() {
		fileName, ok := FileNameForKind(kind)
		if !ok {
			continue
		}
		path := filepath.Join(memoriesDir, fileName)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		parsed, err := ParseMarkdownBlocks(string(data), kind, projectID)
		if err != nil {
			return nil, fmt.Errorf("parse %s (%s): %w", fileName, kind, err)
		}
		records = append(records, parsed...)
	}
	return records, nil
}
