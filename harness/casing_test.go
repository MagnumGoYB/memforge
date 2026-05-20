package harness_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type casingRule struct {
	canonical string
	wrong     []string
	exempt    []string
}

func loadCasingRules(t *testing.T) []casingRule {
	t.Helper()
	path := filepath.Join(repoRoot(t), "docs", "casing-rules.txt")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open casing rules: %v", err)
	}
	defer f.Close()

	var rules []casingRule
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}
		r := casingRule{canonical: strings.TrimSpace(parts[0])}
		for _, w := range strings.Split(parts[1], ",") {
			w = strings.TrimSpace(w)
			if w != "" {
				r.wrong = append(r.wrong, w)
			}
		}
		if len(parts) >= 3 {
			for _, e := range strings.Split(parts[2], ",") {
				e = strings.TrimSpace(e)
				if e != "" {
					r.exempt = append(r.exempt, e)
				}
			}
		}
		rules = append(rules, r)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read casing rules: %v", err)
	}
	return rules
}

var (
	fenceRe      = regexp.MustCompile("^\\s*```")
	inlineCodeRe = regexp.MustCompile("`([^`]+)`")
)

func TestDocumentationCasingConsistency(t *testing.T) {
	rules := loadCasingRules(t)
	root := repoRoot(t)

	var mdFiles []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if info.IsDir() {
			if shouldSkipCasingDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	for _, path := range mdFiles {
		rel, _ := filepath.Rel(root, path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}

		inFence := false
		for lineNo, line := range strings.Split(string(data), "\n") {
			if fenceRe.MatchString(line) {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}

			// Separate prose from inline code spans.
			spans := inlineCodeRe.FindAllStringSubmatch(line, -1)
			proseText := inlineCodeRe.ReplaceAllString(line, "")

			for _, rule := range rules {
				for _, wrong := range rule.wrong {
					if containsWord(proseText, wrong) && !isExemptContext(line, wrong, rule.exempt) {
						t.Errorf("%s:%d: use %q instead of %q", rel, lineNo+1, rule.canonical, wrong)
					}
				}
				for _, m := range spans {
					span := m[1]
					for _, wrong := range rule.wrong {
						if containsWord(span, wrong) && !isExempt(span, rule.exempt) {
							t.Errorf("%s:%d: in inline code, use %q instead of %q", rel, lineNo+1, rule.canonical, wrong)
						}
					}
				}
			}
		}
	}
}

func shouldSkipCasingDir(rel string) bool {
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return false
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		switch part {
		case ".cache", "vendor", "node_modules", ".git":
			return true
		}
	}
	return false
}

func containsWord(text, target string) bool {
	idx := 0
	for {
		i := strings.Index(text[idx:], target)
		if i < 0 {
			return false
		}
		start := idx + i
		end := start + len(target)

		if !strings.Contains(target, " ") {
			validStart := start == 0 || !isWordChar(text[start-1])
			validEnd := end >= len(text) || !isWordChar(text[end])
			if validStart && validEnd {
				return true
			}
		} else {
			return true
		}
		idx = start + 1
		if idx >= len(text) {
			return false
		}
	}
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func isExempt(span string, exemptions []string) bool {
	for _, ex := range exemptions {
		if strings.Contains(span, ex) {
			return true
		}
	}
	return false
}

// isExemptContext checks if every occurrence of wrong in the full line
// is adjacent to an exempt pattern (e.g. "github" in "github-automation.md").
func isExemptContext(fullLine, wrong string, exemptions []string) bool {
	idx := 0
	found := false
	for {
		i := strings.Index(fullLine[idx:], wrong)
		if i < 0 {
			break
		}
		pos := idx + i
		// Extract a window around the match for exempt checking.
		start := pos - 20
		if start < 0 {
			start = 0
		}
		end := pos + len(wrong) + 20
		if end > len(fullLine) {
			end = len(fullLine)
		}
		window := fullLine[start:end]
		if !isExempt(window, exemptions) {
			return false
		}
		found = true
		idx = pos + 1
		if idx >= len(fullLine) {
			break
		}
	}
	return found
}
