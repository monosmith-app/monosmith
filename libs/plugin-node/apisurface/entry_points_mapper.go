package apisurface

import (
	"encoding/json"
	"sort"
	"strings"
)

// EntryPoint is one way in to a package: the main entry or a subpath export,
// with the files that can stand for it, best first.
type EntryPoint struct {
	// Subpath is "." for the main entry, else the exports key ("./cli").
	Subpath    string
	Candidates []string
}

// conditionOrder is the order the conditions of an exports object are tried
// in: the type declarations first, as they are the public contract.
var conditionOrder = []string{"types", "import", "require", "node", "default"}

// EntryPoints maps the text of a package.json to its entry points. Candidates
// are paths relative to the package directory without a leading "./". A
// wildcard subpath ("./*") has no fixed file and is left out. A private
// package has no entry points.
func EntryPoints(manifest string) (points []EntryPoint, private bool, err error) {
	var m struct {
		Private bool            `json:"private"`
		Main    string          `json:"main"`
		Module  string          `json:"module"`
		Types   string          `json:"types"`
		Typings string          `json:"typings"`
		Exports json.RawMessage `json:"exports"`
	}
	if err := json.Unmarshal([]byte(manifest), &m); err != nil {
		return nil, false, err
	}
	if m.Private {
		return nil, true, nil
	}
	fallback := nonEmpty(m.Types, m.Typings, m.Module, m.Main)
	if len(m.Exports) == 0 || string(m.Exports) == "null" {
		return []EntryPoint{{Subpath: ".", Candidates: tidy(append(fallback, "index.d.ts", "index.js"))}}, false, nil
	}
	var exports any
	if err := json.Unmarshal(m.Exports, &exports); err != nil {
		return nil, false, err
	}
	subpaths := map[string]any{}
	if obj, ok := exports.(map[string]any); ok && hasSubpathKeys(obj) {
		for k, v := range obj {
			subpaths[k] = v
		}
	} else {
		subpaths["."] = exports
	}
	keys := make([]string, 0, len(subpaths))
	for k := range subpaths {
		if !strings.Contains(k, "*") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		files := flatten(subpaths[k])
		if k == "." {
			files = append(files, fallback...)
		}
		if c := tidy(files); len(c) > 0 {
			points = append(points, EntryPoint{Subpath: k, Candidates: c})
		}
	}
	return points, false, nil
}

// hasSubpathKeys is whether an exports object maps subpaths (keys that start
// with a dot) instead of conditions.
func hasSubpathKeys(obj map[string]any) bool {
	for k := range obj {
		if strings.HasPrefix(k, ".") {
			return true
		}
	}
	return false
}

// flatten lists the files an exports value can lead to, trying the
// conditions in conditionOrder and then the others in name order.
func flatten(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, item := range t {
			out = append(out, flatten(item)...)
		}
		return out
	case map[string]any:
		keys := make([]string, 0, len(t))
		for _, k := range conditionOrder {
			if _, ok := t[k]; ok {
				keys = append(keys, k)
			}
		}
		var rest []string
		for k := range t {
			if !containsString(conditionOrder, k) {
				rest = append(rest, k)
			}
		}
		sort.Strings(rest)
		var out []string
		for _, k := range append(keys, rest...) {
			out = append(out, flatten(t[k])...)
		}
		return out
	}
	return nil
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func nonEmpty(values ...string) []string {
	var out []string
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// tidy cleans the candidate paths ("./dist/x.js" becomes "dist/x.js"), drops
// wildcards and repeats, and keeps the order.
func tidy(files []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, f := range files {
		f = strings.TrimPrefix(strings.ReplaceAll(f, "\\", "/"), "./")
		if f == "" || strings.Contains(f, "*") || seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}
