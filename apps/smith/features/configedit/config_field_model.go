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

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
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
	// Reason is one line saying why the default is what it is, from the
	// schema's x-reason; empty when the schema gives none.
	Reason string `json:"reason,omitempty"`
	// Options are the values of Enum, each with what it does and whether it
	// is the default, so an editor shows a picker; empty when the setting
	// takes any value of its type.
	Options []Option `json:"options,omitempty"`
	// DeclaredOnly is true for a project setting that only a project declared
	// in smith.toml takes (root, language, kind): a project with a manifest
	// gets them from the manifest and its plugin.
	DeclaredOnly bool `json:"declaredOnly,omitempty"`
}

// Option is one value a setting with a fixed set of values takes.
type Option struct {
	// Value is the word written in smith.toml.
	Value string `json:"value"`
	// Description is one line saying what the value does, from the schema's
	// enumDescriptions.
	Description string `json:"description"`
	// Default is true for the value (or, for a list, each value) the setting
	// takes when smith.toml does not set it.
	Default bool `json:"default,omitempty"`
}

// Table is the table of smith.toml the setting sits in: the path up to its
// last dot ("ci", or "projects.web" for a project's setting).
func (f Field) Table() string { return f.Path[:strings.LastIndex(f.Path, ".")] }

// Key is the setting's name inside its table.
func (f Field) Key() string { return f.Path[strings.LastIndex(f.Path, ".")+1:] }

type schemaNode struct {
	Type                 string                `json:"type"`
	Description          string                `json:"description"`
	Reason               string                `json:"x-reason"`
	Properties           map[string]schemaNode `json:"properties"`
	Items                *schemaNode           `json:"items"`
	Enum                 []string              `json:"enum"`
	EnumDescriptions     []string              `json:"enumDescriptions"`
	Default              json.RawMessage       `json:"default"`
	Minimum              *int                  `json:"minimum"`
	MinLength            int                   `json:"minLength"`
	AdditionalProperties json.RawMessage       `json:"additionalProperties"`
}

// Fields are the settings of smith.toml's fixed tables ([workspace], [ci], [release], [cache],
// [plugins], [agent], [dependencies]), in schema order of tables and by name
// within one. Tables of named entries (stacks, tasks, projects, ...) are
// edited in the file.
func Fields() ([]Field, error) {
	var root schemaNode
	if err := json.Unmarshal(workspace.Schema, &root); err != nil {
		return nil, err
	}
	var out []Field
	for _, table := range []string{"workspace", "ci", "release", "cache", "plugins", "agent", "dependencies"} {
		node := root.Properties[table]
		keys := make([]string, 0, len(node.Properties))
		for k := range node.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			// A table, or a list of tables ([[release.channels]]), is edited in the file.
			if f, ok := fieldOf(table+"."+k, node.Properties[k]); ok {
				out = append(out, f)
			}
		}
	}
	return out, nil
}

// fieldOf is the setting at path that the schema node p describes, with its
// options when it takes a fixed set of values. ok is false for a table or a
// list of tables.
func fieldOf(path string, p schemaNode) (Field, bool) {
	if p.Type == "object" || (p.Type == "array" && p.Items != nil && p.Items.Type == "object") {
		return Field{}, false
	}
	f := Field{Path: path, Type: p.Type, Description: p.Description, Reason: p.Reason, Minimum: p.Minimum, MinLength: p.MinLength}
	descriptions := p.EnumDescriptions
	if p.Items != nil {
		f.Enum, descriptions = p.Items.Enum, p.Items.EnumDescriptions
	}
	if len(p.Enum) > 0 {
		// A string that takes one of a few words, with or without "type": "string".
		f.Type, f.Enum, descriptions = "string", p.Enum, p.EnumDescriptions
	}
	f.Options = options(f.Enum, descriptions, p.Default)
	return f, true
}

// options pairs each value with its description and marks the default: one
// word, or each word of a list.
func options(values, descriptions []string, def json.RawMessage) []Option {
	if len(values) == 0 {
		return nil
	}
	var defaults []string
	var one string
	if json.Unmarshal(def, &one) == nil {
		defaults = []string{one}
	} else {
		_ = json.Unmarshal(def, &defaults) // no default, or not a list of words: none is marked
	}
	out := make([]Option, 0, len(values))
	for i, v := range values {
		o := Option{Value: v, Default: slices.Contains(defaults, v)}
		if i < len(descriptions) {
			o.Description = descriptions[i]
		}
		out = append(out, o)
	}
	return out
}

// DefaultStackPrefix starts the path of a default stack: default_stacks.node
// is the stack Node projects take when they record none.
const DefaultStackPrefix = "default_stacks."

// projectsPrefix starts the path of a project's setting:
// projects.<name>.<key>.
const projectsPrefix = "projects."

// declaredOnlyKeys are the project settings only a project declared in
// smith.toml takes; the validator of declared projects refuses them alone.
var declaredOnlyKeys = []string{"root", "language", "kind"}

// ProjectPath is the path of a project's setting: projects.<name>.<key>, the
// name quoted when it is not a bare TOML key.
//
// Parameters: name is the project's name; key is the setting (tag_prefix).
// Returns: the path Find, Get, Set and Unset take.
func ProjectPath(name, key string) string { return projectsPrefix + TableKey(name) + "." + key }

// ProjectFields are the settings of [projects.<name>] that are one value or a
// list of values, as the schema describes them, by name. Tables inside it
// (release, docker, tasks, ...) are edited in the file.
//
// Parameters: name is the project's name.
// Returns: the fields, with paths from ProjectPath; an error when the
// embedded schema cannot be read.
func ProjectFields(name string) ([]Field, error) {
	var root schemaNode
	if err := json.Unmarshal(workspace.Schema, &root); err != nil {
		return nil, err
	}
	var node schemaNode
	if err := json.Unmarshal(root.Properties["projects"].AdditionalProperties, &node); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(node.Properties))
	for k := range node.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []Field
	for _, k := range keys {
		if f, ok := fieldOf(ProjectPath(name, k), node.Properties[k]); ok {
			f.DeclaredOnly = slices.Contains(declaredOnlyKeys, k)
			out = append(out, f)
		}
	}
	return out, nil
}

// splitProjectPath splits "<name>.<key>" (the path after "projects.") at its
// last dot, and unquotes a quoted name.
func splitProjectPath(rest string) (name, key string, ok bool) {
	i := strings.LastIndex(rest, ".")
	if i <= 0 || i == len(rest)-1 {
		return "", "", false
	}
	name, key = rest[:i], rest[i+1:]
	if unquoted, err := strconv.Unquote(name); err == nil {
		name = unquoted
	}
	return name, key, name != ""
}

// findProjectField is Find for projects.<name>.<key>.
func findProjectField(path string) (Field, error) {
	name, key, ok := splitProjectPath(strings.TrimPrefix(path, projectsPrefix))
	if !ok {
		return Field{}, fmt.Errorf("%s is not a project setting: write projects.<project>.<setting>, such as projects.web.tag_prefix", path)
	}
	fields, err := ProjectFields(name)
	if err != nil {
		return Field{}, err
	}
	var keys []string
	for _, f := range fields {
		if f.Key() == key {
			return f, nil
		}
		keys = append(keys, f.Key())
	}
	return Field{}, fmt.Errorf("%s is not a setting of [projects.%s] that smith config edits. Its settings are %s; tables such as release and tasks are edited in smith.toml", key, TableKey(name), strings.Join(keys, ", "))
}

// Find returns the field at path.
func Find(path string) (Field, error) {
	if language := strings.TrimPrefix(path, DefaultStackPrefix); language != path && language != "" {
		return DefaultStackField(language), nil
	}
	if strings.HasPrefix(path, projectsPrefix) {
		return findProjectField(path)
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
	if strings.HasPrefix(path, "release.channels") {
		return Field{}, errors.New("release channels are a list of tables, which smith config set does not edit: add a [[release.channels]] table to smith.toml (branch, prerelease, major, dist_tag). docs/reference/releasing.md, Release channels, shows one")
	}
	return Field{}, errors.New(message + " Run smith config list to see them, or smith config list -p <project>.")
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
		if len(f.Enum) > 0 && !slices.Contains(f.Enum, args[0]) {
			return "", fmt.Errorf("%s takes %s, not %q", f.Path, strings.Join(f.Enum, " or "), args[0])
		}
		if err := checkNpmScope(f.Path, args[0]); err != nil {
			return "", err
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
