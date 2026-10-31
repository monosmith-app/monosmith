package languageserver

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// SchemaNode is one setting or table of smith.toml, read from the JSON Schema.
type SchemaNode struct {
	// Name is the key, empty for the root.
	Name string
	// Description is the schema's text for the key.
	Description string
	// Type is the JSON Schema type: string, boolean, array, object and so on.
	Type string
	// Enum lists the allowed string values.
	Enum []string
	// Props are the named keys of a table.
	Props map[string]*SchemaNode
	// Additional describes the keys a table takes under any name (a project).
	Additional *SchemaNode
	// Items describes the elements of an array.
	Items *SchemaNode
}

// rawSchema is the part of JSON Schema draft 7 the model reads.
type rawSchema struct {
	Type                 json.RawMessage       `json:"type"`
	Description          string                `json:"description"`
	Enum                 []any                 `json:"enum"`
	Properties           map[string]*rawSchema `json:"properties"`
	AdditionalProperties json.RawMessage       `json:"additionalProperties"`
	Items                *rawSchema            `json:"items"`
	Ref                  string                `json:"$ref"`
	Definitions          map[string]*rawSchema `json:"definitions"`
}

var (
	schemaOnce sync.Once
	schemaRoot *SchemaNode
	schemaErr  error
)

// LoadSchema returns the root of the smith.toml schema embedded in the
// workspace package. It parses the schema once.
func LoadSchema() (*SchemaNode, error) {
	schemaOnce.Do(func() {
		var raw rawSchema
		if schemaErr = json.Unmarshal(workspace.Schema, &raw); schemaErr != nil {
			return
		}
		schemaRoot = buildNode(&raw, raw.Definitions, "", 0)
	})
	return schemaRoot, schemaErr
}

// maxSchemaDepth stops a $ref cycle.
const maxSchemaDepth = 12

func buildNode(raw *rawSchema, defs map[string]*rawSchema, name string, depth int) *SchemaNode {
	if raw == nil || depth > maxSchemaDepth {
		return &SchemaNode{Name: name}
	}
	if raw.Ref != "" {
		node := buildNode(defs[strings.TrimPrefix(raw.Ref, "#/definitions/")], defs, name, depth+1)
		if raw.Description != "" {
			node.Description = raw.Description
		}
		return node
	}
	node := &SchemaNode{Name: name, Description: raw.Description}
	if len(raw.Type) > 0 && raw.Type[0] == '"' {
		_ = json.Unmarshal(raw.Type, &node.Type)
	}
	for _, v := range raw.Enum {
		if s, ok := v.(string); ok {
			node.Enum = append(node.Enum, s)
		}
	}
	if len(raw.Properties) > 0 {
		node.Props = make(map[string]*SchemaNode, len(raw.Properties))
		for key, child := range raw.Properties {
			node.Props[key] = buildNode(child, defs, key, depth+1)
		}
	}
	if len(raw.AdditionalProperties) > 0 && raw.AdditionalProperties[0] == '{' {
		var extra rawSchema
		if json.Unmarshal(raw.AdditionalProperties, &extra) == nil {
			node.Additional = buildNode(&extra, defs, "", depth+1)
		}
	}
	if raw.Items != nil {
		node.Items = buildNode(raw.Items, defs, name, depth+1)
	}
	return node
}

// element is the node whose keys a table header names: an array of tables
// ([[projects.x.y]]) is described by its items.
func (n *SchemaNode) element() *SchemaNode {
	if n.Items != nil && len(n.Props) == 0 && n.Additional == nil {
		return n.Items
	}
	return n
}

// IsTable reports whether the node holds keys.
func (n *SchemaNode) IsTable() bool {
	e := n.element()
	return len(e.Props) > 0 || e.Additional != nil || e.Type == "object"
}

// PropNames returns the named keys of the node, sorted.
func (n *SchemaNode) PropNames() []string {
	e := n.element()
	names := make([]string, 0, len(e.Props))
	for name := range e.Props {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Resolve follows path from the node and returns the node it ends at, or nil
// when the schema has no such key. An empty path returns the node itself.
func (n *SchemaNode) Resolve(path []string) *SchemaNode {
	cur := n
	for _, seg := range path {
		e := cur.element()
		if next := e.Props[seg]; next != nil {
			cur = next
		} else if e.Additional != nil {
			cur = e.Additional
		} else {
			return nil
		}
	}
	return cur
}
