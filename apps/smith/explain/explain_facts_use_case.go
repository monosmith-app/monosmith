package explain

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// FactLine is one shared fact of a project: its value and where it came from.
type FactLine struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
	// Scope is the scope that supplied the value: project, workspace, recipe
	// or plugin default.
	Scope string `json:"scope"`
	// Source names the table or plugin that supplied the value.
	Source string `json:"source"`
}

// ExplainFacts lists the shared facts of a project, sorted by key. With key
// set it lists only that fact. It returns an error when the project does not
// exist, or when key is set and no one set it.
func ExplainFacts(g *projectgraph.Graph, name, key string) ([]FactLine, error) {
	p := g.Get(name)
	if p == nil {
		return nil, fmt.Errorf("no project named %q (projects: %s)", name, strings.Join(g.Names(), ", "))
	}
	keys := make([]string, 0, len(p.FactDetail))
	for k := range p.FactDetail {
		if key == "" || k == key {
			keys = append(keys, k)
		}
	}
	if key != "" && len(keys) == 0 {
		return nil, fmt.Errorf("%s has no fact %q: no plugin published it and smith.toml does not set it (smith explain %s --facts lists them)", name, key, name)
	}
	sort.Strings(keys)
	lines := make([]FactLine, len(keys))
	for i, k := range keys {
		f := p.FactDetail[k]
		lines[i] = FactLine{Key: k, Value: f.Value, Scope: f.Scope.String(), Source: f.Source}
	}
	return lines, nil
}

// valueText shows a fact value on one line, as JSON.
func valueText(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}
