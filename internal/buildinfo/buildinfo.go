// Package buildinfo exposes build metadata wired via -ldflags.
//
// Production binaries set Tag/Commit/Date via the build pipeline; local
// `go build` falls back to the contents of the repository VERSION file.
package buildinfo

import (
	"os"
	"strings"
)

// Overridden via -ldflags at build time.
var (
	Tag    = ""
	Commit = ""
	Date   = ""
)

// Version returns a printable build version string.
//
// Precedence: ldflags Tag > VERSION file > "dev".
func Version() string {
	if Tag != "" {
		return Tag
	}
	if data, err := os.ReadFile("VERSION"); err == nil {
		v := strings.TrimSpace(string(data))
		if v != "" {
			return v
		}
	}
	return "dev"
}
