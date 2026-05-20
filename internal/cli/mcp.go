package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MagnumGOYB/memforge/internal/compiler"
	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/MagnumGOYB/memforge/internal/index"
	"github.com/MagnumGOYB/memforge/internal/mcp"
	"github.com/MagnumGOYB/memforge/internal/memory"
	"github.com/MagnumGOYB/memforge/internal/project"
	"github.com/spf13/cobra"
)

func newMCPCmd(streams Streams) *cobra.Command {
	var rootOverride string
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run the stdio MCP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if _, err := baseconfig.LoadBase(cmd); err != nil {
				return invalidError("%v", err)
			}
			storageRoot, err := baseconfig.ResolveStorageRoot()
			if err != nil {
				return userError("%v", err)
			}
			proj, err := project.Detect(rootOverride)
			if err != nil {
				return userError("%v", err)
			}
			server := newProjectMCPServer(derivePaths(storageRoot, proj), proj)
			return internalError(server.Serve(context.Background(), streams.Stdin, streams.Stdout))
		},
	}
	cmd.Flags().StringVar(&rootOverride, "root", "", "project root override")
	return cmd
}

func newProjectMCPServer(paths resolvedPaths, proj project.Project) mcp.Server {
	tools := []mcp.Tool{
		{Name: "search_memory", Description: "Search project memories", InputSchema: objectSchema(map[string]any{"query": stringSchema(), "kinds": arraySchema(stringSchema()), "limit": integerSchema(), "hybrid": booleanSchema()}, []string{"query"})},
		{Name: "compile_context", Description: "Compile agent-ready project context", InputSchema: objectSchema(map[string]any{"budget": integerSchema(), "kinds": arraySchema(stringSchema())}, nil)},
		{Name: "save_memory", Description: "Persist a project memory", InputSchema: objectSchema(map[string]any{"kind": stringSchema(), "title": stringSchema(), "content": stringSchema(), "tags": arraySchema(stringSchema())}, []string{"kind", "title", "content"})},
		{Name: "upsert_project_memory", Description: "Create or update a stable project memory by kind and title", InputSchema: objectSchema(map[string]any{"kind": stringSchema(), "title": stringSchema(), "content": stringSchema(), "tags": arraySchema(stringSchema())}, []string{"kind", "title", "content"})},
		{Name: "list_constraints", Description: "List high-priority constraint memories", InputSchema: objectSchema(map[string]any{"limit": integerSchema()}, nil)},
		{Name: "get_project_context", Description: "Compile project context, optionally conditioned on a task", InputSchema: objectSchema(map[string]any{"task": stringSchema(), "budget": integerSchema()}, nil)},
	}
	handlers := map[string]mcp.Handler{
		"search_memory":         handleMCPSearch(paths, proj),
		"compile_context":       handleMCPCompile(paths, proj),
		"save_memory":           handleMCPSave(paths, proj),
		"upsert_project_memory": handleMCPUpsert(paths, proj),
		"list_constraints":      handleMCPListConstraints(paths, proj),
		"get_project_context":   handleMCPProjectContext(paths, proj),
	}
	return mcp.Server{Tools: tools, Handlers: handlers}
}

func handleMCPSearch(paths resolvedPaths, proj project.Project) mcp.Handler {
	return func(ctx context.Context, raw json.RawMessage) (any, error) {
		var args struct {
			Query  string   `json:"query"`
			Kinds  []string `json:"kinds"`
			Limit  int      `json:"limit"`
			Hybrid bool     `json:"hybrid"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, err
		}
		if strings.TrimSpace(args.Query) == "" {
			return nil, fmt.Errorf("query is required")
		}
		db, err := index.Open(paths.Index)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		results, err := index.SearchMemories(ctx, db, index.SearchQuery{ProjectID: proj.ID, Query: args.Query, Kinds: args.Kinds, Limit: args.Limit, Hybrid: args.Hybrid})
		if err != nil {
			return nil, err
		}
		return map[string]any{"query": args.Query, "count": len(results), "results": results}, nil
	}
}

func handleMCPCompile(paths resolvedPaths, proj project.Project) mcp.Handler {
	return func(ctx context.Context, raw json.RawMessage) (any, error) {
		var args struct {
			Budget int      `json:"budget"`
			Kinds  []string `json:"kinds"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, err
		}
		records, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
		if err != nil {
			return nil, err
		}
		projectSettings, kindWeights, err := loadCompileSettings(proj.Root)
		if err != nil {
			return nil, err
		}
		budget := resolveBudget(args.Budget, projectSettings)
		kinds, err := parseKinds(args.Kinds)
		if err != nil {
			return nil, err
		}
		result := compiler.CompileContext(compiler.CompileInput{Memories: records, Budget: budget, Kinds: kinds, KindWeights: kindWeights})
		return map[string]any{"budget": budget, "count": len(result.Entries), "context": result.Markdown, "warnings": result.Warnings}, nil
	}
}

func handleMCPSave(paths resolvedPaths, proj project.Project) mcp.Handler {
	return func(ctx context.Context, raw json.RawMessage) (any, error) {
		var args struct {
			Kind    string   `json:"kind"`
			Title   string   `json:"title"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, err
		}
		kind, err := memory.ParseKind(args.Kind)
		if err != nil {
			return nil, err
		}
		record, err := memory.NewRecord(memory.NewRecordInput{ProjectID: proj.ID, Kind: kind, Title: args.Title, Content: args.Content, Tags: args.Tags, Source: "mcp", Confidence: 1})
		if err != nil {
			return nil, err
		}
		if err := persistMemory(ctx, paths, record); err != nil {
			return nil, err
		}
		return map[string]any{"id": record.ID, "kind": record.Kind, "title": record.Title}, nil
	}
}

func handleMCPUpsert(paths resolvedPaths, proj project.Project) mcp.Handler {
	return func(ctx context.Context, raw json.RawMessage) (any, error) {
		var args struct {
			Kind    string   `json:"kind"`
			Title   string   `json:"title"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, err
		}
		kind, err := memory.ParseKind(args.Kind)
		if err != nil {
			return nil, err
		}
		incoming, err := memory.NewRecord(memory.NewRecordInput{ProjectID: proj.ID, Kind: kind, Title: args.Title, Content: args.Content, Tags: args.Tags, Source: "mcp", Confidence: 1})
		if err != nil {
			return nil, err
		}
		records, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
		if err != nil {
			return nil, err
		}
		kindRecords := make([]memory.Record, 0)
		action := "created"
		record := incoming
		for _, existing := range records {
			if existing.Kind != kind {
				continue
			}
			if normalizeMemoryTitle(existing.Title) == normalizeMemoryTitle(incoming.Title) {
				incoming.ID = existing.ID
				incoming.CreatedAt = existing.CreatedAt
				record = incoming
				action = "updated"
				kindRecords = append(kindRecords, incoming)
				continue
			}
			kindRecords = append(kindRecords, existing)
		}
		if action == "created" {
			kindRecords = append(kindRecords, incoming)
		}
		if _, err := memory.RewriteKindMarkdown(paths.MemoriesDir, kind, kindRecords); err != nil {
			return nil, err
		}
		db, err := index.Open(paths.Index)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		indexRecords, err := loadIndexRecords(paths.MemoriesDir, proj.ID)
		if err != nil {
			return nil, err
		}
		if _, err := index.RebuildMemories(ctx, db, indexRecords); err != nil {
			return nil, err
		}
		return map[string]any{"action": action, "id": record.ID, "kind": record.Kind, "title": record.Title}, nil
	}
}

func normalizeMemoryTitle(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(title)), " "))
}

func loadIndexRecords(memoriesDir string, projectID string) ([]index.MemoryRecord, error) {
	records, err := memory.LoadRecords(memoriesDir, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]index.MemoryRecord, 0, len(records))
	for _, record := range records {
		out = append(out, indexRecord(record))
	}
	return out, nil
}

func handleMCPListConstraints(paths resolvedPaths, proj project.Project) mcp.Handler {
	return func(ctx context.Context, raw json.RawMessage) (any, error) {
		var args struct {
			Limit int `json:"limit"`
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &args); err != nil {
				return nil, err
			}
		}
		records, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
		if err != nil {
			return nil, err
		}
		limit := args.Limit
		if limit <= 0 {
			limit = 20
		}
		constraints := make([]memory.Record, 0)
		for _, record := range records {
			if record.Kind != memory.KindConstraint {
				continue
			}
			constraints = append(constraints, record)
			if len(constraints) == limit {
				break
			}
		}
		return map[string]any{"count": len(constraints), "results": constraints}, nil
	}
}

func handleMCPProjectContext(paths resolvedPaths, proj project.Project) mcp.Handler {
	return func(ctx context.Context, raw json.RawMessage) (any, error) {
		var args struct {
			Task   string `json:"task"`
			Budget int    `json:"budget"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, err
		}
		records, err := memory.LoadRecords(paths.MemoriesDir, proj.ID)
		if err != nil {
			return nil, err
		}
		projectSettings, kindWeights, err := loadCompileSettings(proj.Root)
		if err != nil {
			return nil, err
		}
		budget := resolveBudget(args.Budget, projectSettings)
		selected := records
		heading := "Project Context"
		if strings.TrimSpace(args.Task) != "" {
			heading = args.Task
			db, err := index.Open(paths.Index)
			if err != nil {
				return nil, err
			}
			defer db.Close()
			matches, err := index.SearchMemories(ctx, db, index.SearchQuery{ProjectID: proj.ID, Query: args.Task, Limit: 20})
			if err != nil && !strings.Contains(err.Error(), "query is required") {
				return nil, err
			}
			selected = selectBeforeRecords(records, matches, args.Task)
		}
		result := compiler.CompileContext(compiler.CompileInput{Memories: selected, Budget: budget, Heading: heading, KindWeights: kindWeights})
		return map[string]any{"task": args.Task, "budget": budget, "count": len(result.Entries), "context": result.Markdown, "warnings": result.Warnings}, nil
	}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema() map[string]any {
	return map[string]any{"type": "string"}
}

func integerSchema() map[string]any {
	return map[string]any{"type": "integer", "minimum": 0}
}

func booleanSchema() map[string]any {
	return map[string]any{"type": "boolean"}
}

func arraySchema(items any) map[string]any {
	return map[string]any{"type": "array", "items": items}
}
