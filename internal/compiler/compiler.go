package compiler

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/MagnumGOYB/memforge/internal/memory"
)

type CompileInput struct {
	Memories []memory.Record
	Budget   int
	Kinds    []memory.Kind
	Heading  string
}

type Entry struct {
	Record memory.Record
	Tokens int
	Score  float64
}

type CompileResult struct {
	Markdown string
	Entries  []Entry
	Warnings []string
}

func CompileContext(input CompileInput) CompileResult {
	budget := input.Budget
	if budget <= 0 {
		budget = 3000
	}
	allowedKinds := map[memory.Kind]struct{}{}
	if len(input.Kinds) > 0 {
		for _, kind := range input.Kinds {
			allowedKinds[kind] = struct{}{}
		}
	}
	entries := make([]Entry, 0, len(input.Memories))
	warnings := []string{}
	now := time.Now().UTC()
	for _, record := range input.Memories {
		if len(allowedKinds) > 0 {
			if _, ok := allowedKinds[record.Kind]; !ok {
				continue
			}
		}
		count := CountTokens(record.Title + "\n" + record.Content)
		if count.UsedFallback && count.WarningMessage != "" {
			warnings = appendIfMissing(warnings, count.WarningMessage)
		}
		entries = append(entries, Entry{Record: record, Tokens: count.Tokens, Score: scoreRecord(record, now)})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].Record.UpdatedAt.After(entries[j].Record.UpdatedAt)
		}
		return entries[i].Score > entries[j].Score
	})
	selected := allocateBudget(entries, budget)
	heading := input.Heading
	if strings.TrimSpace(heading) == "" {
		heading = "Project Context"
	}
	return CompileResult{Markdown: renderContext(heading, selected), Entries: selected, Warnings: warnings}
}

func allocateBudget(entries []Entry, budget int) []Entry {
	if budget <= 0 || len(entries) == 0 {
		return nil
	}
	selected := make([]Entry, 0, len(entries))
	used := 0
	seen := map[string]struct{}{}
	priorityBudget := int(math.Floor(float64(budget) * 0.30))
	if priorityBudget < budget && priorityBudget == 0 {
		priorityBudget = budget
	}
	priorityKinds := map[memory.Kind]struct{}{
		memory.KindManual:     {},
		memory.KindConstraint: {},
	}
	for _, entry := range entries {
		if _, ok := priorityKinds[entry.Record.Kind]; !ok {
			continue
		}
		if used+entry.Tokens > priorityBudget && len(selected) > 0 {
			continue
		}
		if used+entry.Tokens > budget {
			continue
		}
		selected = append(selected, entry)
		used += entry.Tokens
		seen[entry.Record.ID] = struct{}{}
	}
	kindBuckets := map[memory.Kind][]Entry{}
	for _, entry := range entries {
		if _, ok := seen[entry.Record.ID]; ok {
			continue
		}
		kindBuckets[entry.Record.Kind] = append(kindBuckets[entry.Record.Kind], entry)
	}
	type quota struct {
		kind      memory.Kind
		tokens    int
		remaining int
	}
	weights := make([]quota, 0, len(kindBuckets))
	totalWeight := 0
	for kind := range kindBuckets {
		weight := defaultKindWeight(kind)
		weights = append(weights, quota{kind: kind, tokens: int(math.Floor(float64(budget*weight) / 505.0)), remaining: weight})
		totalWeight += weight
	}
	_ = totalWeight
	sort.SliceStable(weights, func(i, j int) bool {
		return defaultKindWeight(weights[i].kind) > defaultKindWeight(weights[j].kind)
	})
	for i := range weights {
		quotaLeft := weights[i].tokens
		for len(kindBuckets[weights[i].kind]) > 0 {
			entry := kindBuckets[weights[i].kind][0]
			if used+entry.Tokens > budget {
				break
			}
			if quotaLeft > 0 && entry.Tokens > quotaLeft && len(selected) > 0 {
				break
			}
			selected = append(selected, entry)
			used += entry.Tokens
			quotaLeft -= entry.Tokens
			kindBuckets[weights[i].kind] = kindBuckets[weights[i].kind][1:]
		}
	}
	for _, weight := range weights {
		for _, entry := range kindBuckets[weight.kind] {
			if used+entry.Tokens > budget {
				continue
			}
			selected = append(selected, entry)
			used += entry.Tokens
		}
	}
	return selected
}

func renderContext(heading string, entries []Entry) string {
	if len(entries) == 0 {
		return fmt.Sprintf("# %s\n\n_No matching memories._\n", heading)
	}
	grouped := map[memory.Kind][]Entry{}
	order := []memory.Kind{}
	for _, entry := range entries {
		if _, ok := grouped[entry.Record.Kind]; !ok {
			order = append(order, entry.Record.Kind)
		}
		grouped[entry.Record.Kind] = append(grouped[entry.Record.Kind], entry)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", heading))
	for _, kind := range order {
		b.WriteString(fmt.Sprintf("## %s\n\n", kind))
		for _, entry := range grouped[kind] {
			b.WriteString(fmt.Sprintf("### %s\n\n", entry.Record.Title))
			if len(entry.Record.Tags) > 0 {
				b.WriteString(fmt.Sprintf("- tags: %s\n", strings.Join(entry.Record.Tags, ", ")))
			}
			if entry.Record.Source != "" {
				b.WriteString(fmt.Sprintf("- source: %s\n", entry.Record.Source))
			}
			b.WriteString(fmt.Sprintf("- confidence: %.3f\n\n", entry.Record.Confidence))
			b.WriteString(entry.Record.Content)
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

func scoreRecord(record memory.Record, now time.Time) float64 {
	kindWeightNorm := float64(defaultKindWeight(record.Kind)) / 100.0
	daysSinceUpdated := now.Sub(record.UpdatedAt).Hours() / 24
	recencyScore := math.Pow(0.5, daysSinceUpdated/30)
	usageScore := 0.0
	effectiveConfidence := memory.EffectiveConfidence(record.Kind, record.Confidence, record.UpdatedAt, now)
	return 0.40*1.0 + 0.20*kindWeightNorm + 0.15*recencyScore + 0.15*usageScore + 0.10*effectiveConfidence
}

func defaultKindWeight(kind memory.Kind) int {
	switch kind {
	case memory.KindManual:
		return 100
	case memory.KindConstraint:
		return 90
	case memory.KindConvention:
		return 80
	case memory.KindDecision:
		return 70
	case memory.KindBugfix:
		return 60
	case memory.KindAPIContract:
		return 55
	case memory.KindAgentInstruction:
		return 50
	default:
		return 50
	}
}

func appendIfMissing(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}
