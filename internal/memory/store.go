package memory

import (
	"os"
	"path/filepath"
	"strings"
)

type AppendResult struct {
	Path string
}

func AppendMarkdown(memoriesDir string, record Record) (AppendResult, error) {
	if err := EnsureLayout(memoriesDir); err != nil {
		return AppendResult{}, err
	}
	fileName, ok := FileNameForKind(record.Kind)
	if !ok {
		return AppendResult{}, os.ErrInvalid
	}
	path := filepath.Join(memoriesDir, fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return AppendResult{}, err
	}
	defer file.Close()
	if _, err := file.WriteString(RenderMarkdownBlock(record)); err != nil {
		return AppendResult{}, err
	}
	return AppendResult{Path: path}, file.Sync()
}

func RewriteKindMarkdown(memoriesDir string, kind Kind, records []Record) (AppendResult, error) {
	if err := EnsureLayout(memoriesDir); err != nil {
		return AppendResult{}, err
	}
	fileName, ok := FileNameForKind(kind)
	if !ok {
		return AppendResult{}, os.ErrInvalid
	}
	path := filepath.Join(memoriesDir, fileName)
	var builder strings.Builder
	for _, record := range records {
		builder.WriteString(RenderMarkdownBlock(record))
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return AppendResult{}, err
	}
	defer file.Close()
	if _, err := file.WriteString(builder.String()); err != nil {
		return AppendResult{}, err
	}
	return AppendResult{Path: path}, file.Sync()
}
