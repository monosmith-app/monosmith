package mcpserver

import (
	"encoding/json"
	"fmt"
	"slices"
)

// ProtocolVersions are the MCP versions the server speaks, newest first.
var ProtocolVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// Outcome is what running a tool's command produced.
type Outcome struct {
	Text    string
	IsError bool
}

// Server answers MCP requests.
type Server struct {
	Tools   []Tool
	Version string
	// Run runs smith with the arguments.
	Run func(args []string) Outcome
}

type request struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Handle answers one message; nil for a notification, which has no answer.
func (s Server) Handle(raw []byte) []byte {
	var req request
	if err := json.Unmarshal(raw, &req); err != nil {
		return encode(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{-32700, "parse error: " + err.Error()}})
	}
	if len(req.ID) == 0 {
		return nil
	}
	reply := func(result any) []byte { return encode(response{JSONRPC: "2.0", ID: req.ID, Result: result}) }
	refuse := func(code int, msg string) []byte {
		return encode(response{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{code, msg}})
	}
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		version := ProtocolVersions[0]
		if slices.Contains(ProtocolVersions, p.ProtocolVersion) {
			version = p.ProtocolVersion
		}
		return reply(map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "smith", "version": s.Version},
			"instructions":    "smith runs this monorepo: projects in several languages, their tasks from a cache, CI and releases. Start with smith_projects and smith_doctor; smith_kinds and smith_stacks list what smith_add_project can create. To help with a recipe, a plugin or a new project, use smith_recipe_check to show the plan before writing, smith_explain to see where a value came from, and smith_plugin_list for what plugins are available.",
		})
	case "ping":
		return reply(map[string]any{})
	case "tools/list":
		var tools []map[string]any
		for _, t := range s.Tools {
			tools = append(tools, map[string]any{
				"name": t.Name, "description": t.Description, "inputSchema": t.InputSchema,
				"annotations": map[string]any{"readOnlyHint": t.ReadOnly, "destructiveHint": !t.ReadOnly && t.Name == "smith_remove_project", "openWorldHint": false},
			})
		}
		return reply(map[string]any{"tools": tools})
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return refuse(-32602, err.Error())
		}
		i := slices.IndexFunc(s.Tools, func(t Tool) bool { return t.Name == p.Name })
		if i < 0 {
			return refuse(-32602, fmt.Sprintf("unknown tool %q", p.Name))
		}
		if p.Arguments == nil {
			p.Arguments = map[string]any{}
		}
		args, err := s.Tools[i].Args(p.Arguments)
		if err != nil {
			return reply(content(err.Error(), true))
		}
		out := s.Run(args)
		return reply(content(out.Text, out.IsError))
	}
	return refuse(-32601, "method not found: "+req.Method)
}

func content(text string, isError bool) map[string]any {
	return map[string]any{"content": []map[string]any{{"type": "text", "text": text}}, "isError": isError}
}

func encode(r response) []byte {
	b, _ := json.Marshal(r)
	return b
}
