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
	matchExpr := buildMatchQuery(query.Query)
	if matchExpr == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
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
	sqlText += ` ORDER BY bm25(memories_fts) LIMIT ?`
	args = append(args, limit)
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
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	scoreResults(results)
	if query.Hybrid {
		rerankHybrid(results, query.Query)
	}
	return results, nil
}

func buildMatchQuery(raw string) string {
	tokens := tokenPattern.FindAllString(strings.ToLower(raw), -1)
	if len(tokens) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, fmt.Sprintf(`"%s"*`, token))
	}
	return strings.Join(parts, " ")
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
			start := idx - 48
			if start < 0 {
				start = 0
			}
			end := idx + len(token) + 96
			if end > len(content) {
				end = len(content)
			}
			snippet := strings.TrimSpace(content[start:end])
			if start > 0 {
				snippet = "..." + snippet
			}
			if end < len(content) {
				snippet += "..."
			}
			return snippet
		}
	}
	if len(content) <= 140 {
		return content
	}
	return content[:140] + "..."
}
