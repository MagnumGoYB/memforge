package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type Handler func(ctx context.Context, arguments json.RawMessage) (any, error)

type Server struct {
	Tools    []Tool
	Handlers map[string]Handler
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id"`
	Result  any            `json:"result,omitempty"`
	Error   *responseError `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	writer := bufio.NewWriter(out)
	defer writer.Flush()
	reader := bufio.NewReader(in)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line, err := reader.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			if err := writeResponse(writer, response{JSONRPC: "2.0", ID: nil, Error: &responseError{Code: -32700, Message: "parse error"}}); err != nil {
				return err
			}
			continue
		}
		if req.ID == nil {
			continue
		}
		resp := s.handle(ctx, req)
		if err := writeResponse(writer, resp); err != nil {
			return err
		}
		if err == io.EOF {
			return nil
		}
	}
}

func (s Server) handle(ctx context.Context, req request) response {
	resp := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "memforge", "version": "dev"},
		}
	case "tools/list":
		resp.Result = map[string]any{"tools": s.Tools}
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &responseError{Code: -32602, Message: "invalid tool call params"}
			return resp
		}
		handler, ok := s.Handlers[params.Name]
		if !ok {
			resp.Error = &responseError{Code: -32601, Message: fmt.Sprintf("unknown tool %q", params.Name)}
			return resp
		}
		result, err := handler(ctx, params.Arguments)
		if err != nil {
			resp.Error = &responseError{Code: -32000, Message: err.Error()}
			return resp
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			resp.Error = &responseError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = map[string]any{"content": []map[string]any{{"type": "text", "text": string(encoded)}}}
	default:
		resp.Error = &responseError{Code: -32601, Message: fmt.Sprintf("unknown method %q", req.Method)}
	}
	return resp
}

func writeResponse(w *bufio.Writer, resp response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}
	return w.Flush()
}
