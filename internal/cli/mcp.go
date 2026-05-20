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
		{Name: "save_memory", Description: "Persist a human-confirmed project memory", InputSchema: objectSchema(map[string]any{"kind": stringSchema(), "title": stringSchema(), "content": stringSchema(), "tags": arraySchema(stringSchema())}, []string{"kind", "title", "content"})},
		{Name: "list_constraints", Description: "List high-priority constraint memories", InputSchema: objectSchema(map[string]any{"limit": integerSchema()}, nil)},
		{Name: "get_project_context", Description: "Compile project context, optionally conditioned on a task", InputSchema: objectSchema(map[string]any{"task": stringSchema(), "budget": integerSchema()}, nil)},
	}
	handlers := map[string]mcp.Handler{
		"search_memory":       handleMCPSearch(paths, proj),
		"compile_context":     handleMCPCompile(paths, proj),
		"save_memory":         handleMCPSave(paths, proj),
		"list_constraints":    handleMCPListConstraints(paths, proj),
		"get_project_context": handleMCPProjectContext(paths, proj),
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
		kinds, err := parseKinds(args.Kinds)
		if err != nil {
			return nil, err
		}
		result := compiler.CompileContext(compiler.CompileInput{Memories: records, Budget: args.Budget, Kinds: kinds})
		return map[string]any{"budget": args.Budget, "count": len(result.Entries), "context": result.Markdown, "warnings": result.Warnings}, nil
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
		_, warning, err := persistMemory(ctx, paths, record)
		if err != nil {
			return nil, err
		}
		payload := map[string]any{"id": record.ID, "kind": record.Kind, "title": record.Title}
		if warning != "" {
			payload["warning"] = warning
		}
		return payload, nil
	}
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
		result := compiler.CompileContext(compiler.CompileInput{Memories: selected, Budget: args.Budget, Heading: heading})
		return map[string]any{"task": args.Task, "count": len(result.Entries), "context": result.Markdown, "warnings": result.Warnings}, nil
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
