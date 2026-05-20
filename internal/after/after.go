package after

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MagnumGOYB/memforge/internal/memory"
)

type Candidate struct {
	ID         string      `json:"id"`
	Kind       memory.Kind `json:"kind"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	Tags       []string    `json:"tags,omitempty"`
	Source     string      `json:"source,omitempty"`
	Confidence float64     `json:"confidence"`
}

type Duplicate struct {
	CandidateID string  `json:"candidate_id"`
	MemoryID    string  `json:"memory_id"`
	Reason      string  `json:"reason"`
	Score       float64 `json:"score"`
}

type MergeProposal struct {
	CandidateID string  `json:"candidate_id"`
	MemoryID    string  `json:"memory_id"`
	Reason      string  `json:"reason"`
	Score       float64 `json:"score"`
}

var kindLinePattern = regexp.MustCompile(`(?i)^\s*(manual|constraint|convention|decision|bugfix|api-contract|agent-instruction)\s*:\s*(.+)$`)

func ExtractCandidatesFromText(text string) []Candidate {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	blocks := strings.Split(text, "\n\n")
	candidates := make([]Candidate, 0)
	for _, block := range blocks {
		lines := compactLines(block)
		if len(lines) == 0 {
			continue
		}
		match := kindLinePattern.FindStringSubmatch(lines[0])
		if len(match) != 3 {
			continue
		}
		kind, err := memory.ParseKind(strings.ToLower(match[1]))
		if err != nil {
			continue
		}
		title := strings.TrimSpace(match[2])
		body := strings.TrimSpace(strings.Join(lines[1:], "\n"))
		if body == "" {
			body = title
		}
		candidate := Candidate{
			ID:         candidateID(len(candidates) + 1),
			Kind:       kind,
			Title:      title,
			Content:    body,
			Tags:       inferTags(title + " " + body),
			Source:     "after",
			Confidence: 0.7,
		}
		if candidate.Title != "" && candidate.Content != "" {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func FindDuplicateCandidates(candidates []Candidate, existing []memory.Record) []Duplicate {
	duplicates := make([]Duplicate, 0)
	for _, candidate := range candidates {
		for _, record := range existing {
			if candidate.Kind != record.Kind {
				continue
			}
			score := similarity(candidate.Title+" "+candidate.Content, record.Title+" "+record.Content)
			if normalizeText(candidate.Title) == normalizeText(record.Title) || score >= 0.92 {
				duplicates = append(duplicates, Duplicate{CandidateID: candidate.ID, MemoryID: record.ID, Reason: "same kind and highly similar content", Score: score})
			}
		}
	}
	return duplicates
}

func BuildMergeProposals(candidates []Candidate, existing []memory.Record) []MergeProposal {
	proposals := make([]MergeProposal, 0)
	for _, candidate := range candidates {
		for _, record := range existing {
			if candidate.Kind != record.Kind {
				continue
			}
			titleScore := similarity(candidate.Title, record.Title)
			tagScore := tagOverlap(candidate.Tags, record.Tags)
			score := titleScore*0.75 + tagScore*0.25
			if score >= 0.55 && normalizeText(candidate.Content) != normalizeText(record.Content) {
				proposals = append(proposals, MergeProposal{CandidateID: candidate.ID, MemoryID: record.ID, Reason: "same kind with related title or tags", Score: score})
			}
		}
	}
	sort.SliceStable(proposals, func(i, j int) bool {
		return proposals[i].Score > proposals[j].Score
	})
	return proposals
}

func ApproveCandidates(candidates []Candidate, approval string, duplicates []Duplicate) []Candidate {
	approval = strings.TrimSpace(approval)
	if approval == "" || approval == "none" {
		return nil
	}
	duplicateIDs := map[string]struct{}{}
	for _, duplicate := range duplicates {
		duplicateIDs[duplicate.CandidateID] = struct{}{}
	}
	approvedIDs := map[string]struct{}{}
	if approval != "all" {
		for _, id := range strings.Split(approval, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				approvedIDs[id] = struct{}{}
			}
		}
	}
	approved := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if _, duplicate := duplicateIDs[candidate.ID]; duplicate {
			continue
		}
		if approval == "all" {
			approved = append(approved, candidate)
			continue
		}
		if _, ok := approvedIDs[candidate.ID]; ok {
			approved = append(approved, candidate)
		}
	}
	return approved
}

func compactLines(block string) []string {
	rawLines := strings.Split(strings.TrimSpace(block), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func candidateID(n int) string {
	return "cand_" + strconv.Itoa(n)
}

func inferTags(text string) []string {
	words := tokenSet(text)
	priority := []string{"architecture", "repository", "auth", "cli", "api", "database", "sqlite", "mcp", "provider", "testing"}
	tags := make([]string, 0, 3)
	for _, word := range priority {
		if _, ok := words[word]; ok {
			tags = append(tags, word)
		}
		if len(tags) == 3 {
			break
		}
	}
	return tags
}

func similarity(left string, right string) float64 {
	leftSet := tokenSet(left)
	rightSet := tokenSet(right)
	if len(leftSet) == 0 || len(rightSet) == 0 {
		return 0
	}
	intersection := 0
	for token := range leftSet {
		if _, ok := rightSet[token]; ok {
			intersection++
		}
	}
	union := len(leftSet) + len(rightSet) - intersection
	return float64(intersection) / float64(union)
}

func tagOverlap(left []string, right []string) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	leftSet := map[string]struct{}{}
	for _, tag := range left {
		leftSet[strings.ToLower(tag)] = struct{}{}
	}
	intersection := 0
	for _, tag := range right {
		if _, ok := leftSet[strings.ToLower(tag)]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(leftSet)+len(right)-intersection)
}

func tokenSet(text string) map[string]struct{} {
	text = normalizeText(text)
	set := map[string]struct{}{}
	for _, token := range strings.Fields(text) {
		if len(token) < 3 {
			continue
		}
		set[token] = struct{}{}
	}
	return set
}

func normalizeText(text string) string {
	text = strings.ToLower(text)
	var b strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(' ')
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
