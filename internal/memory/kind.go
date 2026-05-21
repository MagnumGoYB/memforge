package memory

import (
	"fmt"
	"strings"
)

type Kind string

const (
	KindManual           Kind = "manual"
	KindConstraint       Kind = "constraint"
	KindConvention       Kind = "convention"
	KindDecision         Kind = "decision"
	KindBugfix           Kind = "bugfix"
	KindAPIContract      Kind = "api-contract"
	KindAgentInstruction Kind = "agent-instruction"
)

var defaultFiles = []struct {
	Kind Kind
	File string
}{
	{Kind: KindManual, File: "manual.md"},
	{Kind: KindConstraint, File: "constraints.md"},
	{Kind: KindConvention, File: "conventions.md"},
	{Kind: KindDecision, File: "decisions.md"},
	{Kind: KindBugfix, File: "bugfixes.md"},
	{Kind: KindAPIContract, File: "api-contracts.md"},
	{Kind: KindAgentInstruction, File: "agent-instructions.md"},
}

var kindAliases = map[string]Kind{
	"workflow": KindAgentInstruction,
	"note":     KindManual,
	"domain":   KindConvention,
}

func DefaultFileNames() []string {
	files := make([]string, 0, len(defaultFiles))
	for _, entry := range defaultFiles {
		files = append(files, entry.File)
	}
	return files
}

func ParseKind(value string) (Kind, error) {
	value = strings.TrimSpace(value)
	if alias, ok := kindAliases[value]; ok {
		return alias, nil
	}
	for _, entry := range defaultFiles {
		if string(entry.Kind) == value {
			return entry.Kind, nil
		}
	}
	return "", fmt.Errorf("invalid kind %q", value)
}

func FileNameForKind(kind Kind) (string, bool) {
	for _, entry := range defaultFiles {
		if entry.Kind == kind {
			return entry.File, true
		}
	}
	return "", false
}

func ValidKinds() []Kind {
	kinds := make([]Kind, 0, len(defaultFiles))
	for _, entry := range defaultFiles {
		kinds = append(kinds, entry.Kind)
	}
	return kinds
}
