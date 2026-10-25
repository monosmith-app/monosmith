package configedit

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Field is one setting of a fixed table of smith.toml, as the schema
// describes it.
type Field struct {
	// Path is "table.key", e.g. "ci.devcontainer".
	Path        string   `json:"path"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Minimum     *int     `json:"minimum,omitempty"`
	MinLength   int      `json:"minLength,omitempty"`
}

// Table and Key split the path.
func (f Field) Table() string { return strings.SplitN(f.Path, ".", 2)[0] }

// Key is the setting's name inside its table.
func (f Field) Key() string { return strings.SplitN(f.Path, ".", 2)[1] }

type schemaNode struct {
	Type                 string                `json:"type"`
	Description          string                `json:"description"`
	Properties           map[string]schemaNode `json:"properties"`
	Items                *schemaNode           `json:"items"`
	Enum                 []string              `json:"enum"`
	Minimum              *int                  `json:"minimum"`
	MinLength            int                   `json:"minLength"`
	AdditionalProperties json.RawMessage       `json:"additionalProperties"`
}

// Fields are the settings of smith.toml's fixed tables ([workspace], [ci],
// [release], [cache], [plugins]), in schema order of tables and by name
// within one. Tables of named entries (stacks, tasks, projects, ...) are
// edited in the file.
func Fields() ([]Field, error) {
	var root schemaNode
	if err := json.Unmarshal(workspace.Schema, &root); err != nil {
		return nil, err
	}
	var out []Field
	for _, table := range []string{"workspace", "ci", "release", "cache", "plugins"} {
		node := root.Properties[table]
		keys := make([]string, 0, len(node.Properties))
		for k := range node.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			p := node.Properties[k]
			if p.Type == "object" {
				continue
			}
			f := Field{Path: table + "." + k, Type: p.Type, Description: p.Description, Minimum: p.Minimum, MinLength: p.MinLength}
			if p.Items != nil {
				f.Enum = p.Items.Enum
			}
			out = append(out, f)
		}
	}
	return out, nil
}

// DefaultStackPrefix starts the path of a default stack: default_stacks.node
// is the stack Node projects take when they record none.
const DefaultStackPrefix = "default_stacks."

// Find returns the field at path.
func Find(path string) (Field, error) {
	if language := strings.TrimPrefix(path, DefaultStackPrefix); language != path && language != "" {
		return DefaultStackField(language), nil
	}
	fields, err := Fields()
	if err != nil {
		return Field{}, err
	}
	var paths []string
	for _, f := range fields {
		if f.Path == path {
			return f, nil
		}
		paths = append(paths, f.Path)
	}
	message := path + " is not a setting smith config edits"
	if near := nearest(path, paths); len(near) > 0 {
		message += "; did you mean " + strings.Join(near, " or ") + "?"
	}
	return Field{}, errors.New(message + " Run smith config list to see them (stacks, tasks and projects are edited in smith.toml).")
}

// DefaultStackField is the setting that names the stack of a language's
// projects that record none.
func DefaultStackField(language string) Field {
	return Field{
		Path:        DefaultStackPrefix + language,
		Type:        "string",
		Description: "the stack of " + language + " projects that record none, and of smith add without --stack (for example node-jest)",
		MinLength:   1,
	}
}

// nearest returns up to two settings whose name is close to path's: the same
// last word, or a few letters away.
func nearest(path string, paths []string) []string {
	last := path[strings.LastIndex(path, ".")+1:]
	var out []string
	for _, p := range paths {
		name := p[strings.LastIndex(p, ".")+1:]
		if name == last || (len(last) > 2 && strings.Contains(p, last)) || editDistance(path, p) <= 3 {
			out = append(out, p)
		}
		if len(out) == 2 {
			break
		}
	}
	return out
}

func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

// splitChoices splits each argument at commas, so github,azure and github azure
// are the same list.
func splitChoices(args []string) []string {
	var out []string
	for _, a := range args {
		for _, part := range strings.Split(a, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// Encode checks args against the field and renders them as a TOML value: one
// arg for a scalar, any number for an array.
func (f Field) Encode(args []string) (string, error) {
	if f.Type != "array" && len(args) != 1 {
		return "", fmt.Errorf("%s takes one value", f.Path)
	}
	switch f.Type {
	case "boolean":
		b, err := strconv.ParseBool(args[0])
		if err != nil {
			return "", fmt.Errorf("%s is true or false, not %q", f.Path, args[0])
		}
		return strconv.FormatBool(b), nil
	case "integer":
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("%s is a whole number, not %q", f.Path, args[0])
		}
		if f.Minimum != nil && n < *f.Minimum {
			return "", fmt.Errorf("%s is at least %d", f.Path, *f.Minimum)
		}
		return strconv.Itoa(n), nil
	case "string":
		if len(args[0]) < f.MinLength {
			return "", fmt.Errorf("%s cannot be empty", f.Path)
		}
		return Quote(args[0]), nil
	case "array":
		if len(f.Enum) > 0 {
			// Choices never contain a comma, so "a,b,c" is three of them.
			args = splitChoices(args)
		}
		quoted := make([]string, 0, len(args))
		var seen []string
		for _, a := range args {
			if len(f.Enum) > 0 && !slices.Contains(f.Enum, a) {
				return "", fmt.Errorf("%s takes %s, not %q", f.Path, strings.Join(f.Enum, " or "), a)
			}
			if len(f.Enum) > 0 && slices.Contains(seen, a) {
				return "", fmt.Errorf("%s lists %s twice", f.Path, a)
			}
			seen = append(seen, a)
			quoted = append(quoted, Quote(a))
		}
		return "[" + strings.Join(quoted, ", ") + "]", nil
	}
	return "", fmt.Errorf("%s has a type smith config does not edit (%s)", f.Path, f.Type)
}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// TableKey is a project's name as a key in a table path, quoted when it is not
// a bare TOML key (ms.sharepoint).
func TableKey(name string) string {
	if bareKey.MatchString(name) {
		return name
	}
	return Quote(name)
}

// Quote renders a TOML basic string.
func Quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`, "\r", `\r`)
	return `"` + r.Replace(s) + `"`
}
