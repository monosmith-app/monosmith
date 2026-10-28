package agentfiles

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Launcher is how a client starts smith mcp.
type Launcher struct {
	Command string
	Args    []string
}

// The two ways smith is started. A smith on the PATH is the usual one (every
// way of installing it globally puts it there); a smith that is a dependency
// of the workspace's own package.json is not on the PATH, and npx finds it.
var (
	DirectLauncher = Launcher{Command: "smith", Args: []string{"mcp"}}
	NpxLauncher    = Launcher{Command: "npx", Args: []string{"smith", "mcp"}}
)

// MCPForm is the shape of one client's MCP file.
type MCPForm struct {
	Path string
	// Container is the top-level key holding the servers.
	Container string
	// Typed adds "type": "stdio" to the entry, which VS Code requires.
	Typed bool
}

// The forms. Claude Code: https://code.claude.com/docs/en/mcp (project scope,
// mcpServers). VS Code: https://code.visualstudio.com/docs/agents/reference/mcp-configuration
// (servers, with type).
var (
	ClaudeForm = MCPForm{Path: ClaudeMCPPath, Container: "mcpServers"}
	VSCodeForm = MCPForm{Path: VSCodeMCPPath, Container: "servers", Typed: true}
)

// ServerEntry is the smith entry of an MCP file.
type ServerEntry struct {
	Type    string   `json:"type,omitempty"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// Entry is the server entry for a launcher.
func (f MCPForm) Entry(l Launcher) ServerEntry {
	e := ServerEntry{Command: l.Command, Args: l.Args}
	if f.Typed {
		e.Type = "stdio"
	}
	return e
}

// ErrNotJSON is a file smith cannot merge into: it is not plain JSON (VS Code
// accepts comments in mcp.json, smith does not edit around them).
var ErrNotJSON = errors.New("not plain JSON")

type member struct {
	key string
	raw json.RawMessage
}

// members reads a JSON object keeping the order of its keys.
func members(raw []byte) ([]member, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, fmt.Errorf("%w: expected an object", ErrNotJSON)
	}
	var out []member
	for dec.More() {
		k, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrNotJSON, err)
		}
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrNotJSON, err)
		}
		out = append(out, member{key: k.(string), raw: v})
	}
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotJSON, err)
	}
	return out, nil
}

func object(ms []member) []byte {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, m := range ms {
		if i > 0 {
			b.WriteByte(',')
		}
		k, _ := json.Marshal(m.key)
		b.Write(k)
		b.WriteByte(':')
		b.Write(m.raw)
	}
	b.WriteByte('}')
	return b.Bytes()
}

func indented(compact []byte) []byte {
	var b bytes.Buffer
	if err := json.Indent(&b, compact, "", "  "); err != nil {
		return compact
	}
	b.WriteByte('\n')
	return b.Bytes()
}

func set(ms []member, key string, raw []byte) []member {
	for i := range ms {
		if ms[i].key == key {
			ms[i].raw = raw
			return ms
		}
	}
	return append(ms, member{key: key, raw: raw})
}

// MCPState says how a client's file stands.
type MCPState int

// The states.
const (
	// MCPMissing: the file or the smith entry is absent.
	MCPMissing MCPState = iota
	// MCPCurrent: the smith entry is one smith writes.
	MCPCurrent
	// MCPEdited: the smith entry says something else, so it is the person's.
	MCPEdited
)

// MergeMCP returns the file with the smith server added and everything else
// kept: other servers, other keys and their order. An entry named smith that
// is not one smith writes is reported MCPEdited, and the file returned is what
// replacing it would make (the caller shows the difference and writes nothing).
func MergeMCP(existing []byte, f MCPForm, l Launcher) ([]byte, MCPState, error) {
	if len(bytes.TrimSpace(existing)) == 0 {
		existing = []byte("{}")
	}
	top, err := members(existing)
	if err != nil {
		return nil, MCPMissing, err
	}
	var servers []member
	for _, m := range top {
		if m.key == f.Container {
			if servers, err = members(m.raw); err != nil {
				return nil, MCPMissing, fmt.Errorf("%q is not an object: %w", f.Container, err)
			}
		}
	}
	state := MCPMissing
	for _, m := range servers {
		if m.key != ServerName {
			continue
		}
		var have any
		if json.Unmarshal(m.raw, &have) == nil {
			for _, known := range []Launcher{DirectLauncher, NpxLauncher} {
				want, _ := json.Marshal(f.Entry(known))
				var w any
				_ = json.Unmarshal(want, &w)
				if reflect.DeepEqual(have, w) {
					return existing, MCPCurrent, nil
				}
			}
		}
		state = MCPEdited
	}
	entry, _ := json.Marshal(f.Entry(l))
	servers = set(servers, ServerName, entry)
	top = set(top, f.Container, object(servers))
	return indented(object(top)), state, nil
}
