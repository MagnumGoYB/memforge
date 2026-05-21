package index

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/MagnumGOYB/memforge/internal/embedding"
	"github.com/MagnumGOYB/memforge/internal/memory"
)

type SearchQuery struct {
	ProjectID string
	Query     string
	Kinds     []string
	Tags      []string
	Limit     int
	Hybrid    bool
}

type SearchResult struct {
	ID         string
	ProjectID  string
	Kind       string
	Title      string
	Content    string
	Tags       []string
	Source     string
	Confidence float64
	UsageCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Snippet    string
	Score      float64
	bm25       float64
}

var tokenPattern = regexp.MustCompile(`[\pL\pN][\pL\pN_-]*`)

func SearchMemories(ctx context.Context, db *sql.DB, query SearchQuery) ([]SearchResult, error) {
	tokens := queryTokens(query.Query)
	matchExpr := buildMatchQueryFromTokens(tokens, " ")
	if matchExpr == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	requiredTags := normalizeTags(query.Tags)
	results, err := searchMemoriesFTS(ctx, db, query, matchExpr, requiredTags, limit)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 && len(tokens) > 1 {
		fallbackExpr := buildMatchQueryFromTokens(tokens, " OR ")
		results, err = searchMemoriesFTS(ctx, db, query, fallbackExpr, requiredTags, limit)
		if err != nil {
			return nil, err
		}
	}
	scoreResults(results)
	if query.Hybrid {
		rerankHybrid(results, query.Query)
	}
	return results, nil
}

func searchMemoriesFTS(ctx context.Context, db *sql.DB, query SearchQuery, matchExpr string, requiredTags []string, limit int) ([]SearchResult, error) {
	args := []any{query.ProjectID, matchExpr}
	sqlText := `SELECT m.id, m.project_id, m.kind, m.title, m.content, m.tags_json, m.source, m.confidence, m.usage_count, m.created_at, m.updated_at, bm25(memories_fts)
	FROM memories_fts
	JOIN memories m ON m.rowid = memories_fts.rowid
	WHERE m.project_id = ? AND memories_fts MATCH ?`
	if len(query.Kinds) > 0 {
		kindPlaceholders := make([]string, 0, len(query.Kinds))
		for _, kind := range query.Kinds {
			kindPlaceholders = append(kindPlaceholders, "?")
			args = append(args, kind)
		}
		sqlText += ` AND m.kind IN (` + strings.Join(kindPlaceholders, ",") + `)`
	}
	if len(requiredTags) > 0 {
		sqlText += ` AND ` + tagFilterSQL(len(requiredTags))
		for _, tag := range requiredTags {
			encodedTag, err := json.Marshal(tag)
			if err != nil {
				return nil, err
			}
			args = append(args, string(encodedTag))
		}
	}
	sqlText += ` ORDER BY bm25(memories_fts)`
	if len(requiredTags) == 0 {
		sqlText += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]SearchResult, 0)
	for rows.Next() {
		var result SearchResult
		var tagsJSON string
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&result.ID, &result.ProjectID, &result.Kind, &result.Title, &result.Content, &tagsJSON, &result.Source, &result.Confidence, &result.UsageCount, &createdAt, &updatedAt, &result.bm25); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(tagsJSON), &result.Tags); err != nil {
			return nil, err
		}
		result.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		result.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, err
		}
		result.Snippet = buildSnippet(result.Content, query.Query)
		results = append(results, result)
		if len(requiredTags) > 0 && len(results) == limit {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func buildMatchQuery(raw string) string {
	return buildMatchQueryFromTokens(queryTokens(raw), " ")
}

func queryTokens(raw string) []string {
	tokens := tokenPattern.FindAllString(strings.ToLower(raw), -1)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func buildMatchQueryFromTokens(tokens []string, sep string) string {
	if len(tokens) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, fmt.Sprintf(`"%s"*`, token))
	}
	return strings.Join(parts, sep)
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func tagFilterSQL(count int) string {
	clauses := make([]string, 0, count)
	for range count {
		clauses = append(clauses, `instr(m.tags_json, ?) > 0`)
	}
	return strings.Join(clauses, ` AND `)
}

func scoreResults(results []SearchResult) {
	if len(results) == 0 {
		return
	}
	minBM25 := results[0].bm25
	maxBM25 := results[0].bm25
	now := time.Now().UTC()
	for _, result := range results {
		if result.bm25 < minBM25 {
			minBM25 = result.bm25
		}
		if result.bm25 > maxBM25 {
			maxBM25 = result.bm25
		}
	}
	for i := range results {
		bm25Norm := 1.0
		if maxBM25 > minBM25 {
			bm25Norm = 1 - ((results[i].bm25 - minBM25) / (maxBM25 - minBM25))
		}
		kindWeightNorm := float64(defaultKindWeight(results[i].Kind)) / 100.0
		daysSinceUpdated := now.Sub(results[i].UpdatedAt).Hours() / 24
		recencyScore := math.Pow(0.5, daysSinceUpdated/30)
		usageScore := math.Min(1, math.Log10(1+float64(results[i].UsageCount))/2)
		effectiveConfidence := memory.EffectiveConfidence(memory.Kind(results[i].Kind), results[i].Confidence, results[i].UpdatedAt, now)
		results[i].Score = 0.40*bm25Norm + 0.20*kindWeightNorm + 0.15*recencyScore + 0.15*usageScore + 0.10*effectiveConfidence
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].UpdatedAt.After(results[j].UpdatedAt)
		}
		return results[i].Score > results[j].Score
	})
}

func rerankHybrid(results []SearchResult, query string) {
	queryVector := embedding.Embed(query)
	for i := range results {
		contentVector := embedding.Embed(results[i].Title + " " + results[i].Content + " " + strings.Join(results[i].Tags, " "))
		semanticScore := (embedding.Cosine(queryVector, contentVector) + 1) / 2
		results[i].Score = 0.70*results[i].Score + 0.30*semanticScore
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].UpdatedAt.After(results[j].UpdatedAt)
		}
		return results[i].Score > results[j].Score
	})
}

func defaultKindWeight(kind string) int {
	switch kind {
	case "manual":
		return 100
	case "constraint":
		return 90
	case "convention":
		return 80
	case "decision":
		return 70
	case "bugfix":
		return 60
	case "api-contract":
		return 55
	case "agent-instruction":
		return 50
	default:
		return 50
	}
}

func buildSnippet(content string, query string) string {
	content = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\n", " "))
	if content == "" {
		return ""
	}
	lowerContent := strings.ToLower(content)
	for _, token := range tokenPattern.FindAllString(strings.ToLower(query), -1) {
		idx := strings.Index(lowerContent, token)
		if idx >= 0 {
			return runeSnippet(content, idx, len(token), 48, 96)
		}
	}
	runes := []rune(content)
	if len(runes) <= 140 {
		return content
	}
	return string(runes[:140]) + "..."
}

func runeSnippet(content string, byteStart int, byteLen int, before int, after int) string {
	runes := []rune(content)
	startRune := 0
	endRune := len(runes)
	currentByte := 0
	matchStartRune := 0
	matchEndRune := len(runes)
	for i, r := range runes {
		if currentByte == byteStart {
			matchStartRune = i
		}
		currentByte += len(string(r))
		if currentByte == byteStart+byteLen {
			matchEndRune = i + 1
			break
		}
	}
	startRune = matchStartRune - before
	if startRune < 0 {
		startRune = 0
	}
	endRune = matchEndRune + after
	if endRune > len(runes) {
		endRune = len(runes)
	}
	snippet := strings.TrimSpace(string(runes[startRune:endRune]))
	if startRune > 0 {
		snippet = "..." + snippet
	}
	if endRune < len(runes) {
		snippet += "..."
	}
	return snippet
}
