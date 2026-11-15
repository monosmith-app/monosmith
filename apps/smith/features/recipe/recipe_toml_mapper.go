package recipe

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// topKeys, blockKeys and choiceKeys are what a recipe, a block and a choice
// accept; anything else is refused with a list to choose from.
var (
	topKeys    = []string{"name", "version", "extends", "choices", "use", "settings", "when"}
	blockKeys  = []string{"choices", "use", "settings", "when"}
	choiceKeys = []string{"ask", "options", "default"}
)

// ParseRecipe reads a recipe file. source is its path or URL, for messages.
// An unknown key, a value of the wrong type, a missing name, a default that
// is not an option and a use entry that is neither a plugin name nor a list
// of them are all refused with the file and line.
func ParseRecipe(data []byte, source Source) (Document, error) {
	var r Recipe
	dec := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return Document{}, decodeError(err, source)
	}
	doc := Document{Recipe: r, Source: source, lines: indexLines(data)}
	if err := checkRecipe(doc); err != nil {
		return Document{}, err
	}
	return doc, nil
}

// decodeError turns go-toml's errors into one that names the file, line and
// column, and for an unknown key the keys that would have been accepted.
func decodeError(err error, source Source) error {
	var strict *toml.StrictMissingError
	if errors.As(err, &strict) && len(strict.Errors) > 0 {
		var lines []string
		for _, e := range strict.Errors {
			row, col := e.Position()
			key := e.Key()
			lines = append(lines, fmt.Sprintf("%s:%d:%d: unknown key %q; here a recipe accepts %s", source.Ref, row, col, strings.Join(key, "."), strings.Join(acceptedKeys(key), ", ")))
		}
		return &ParseError{Source: source.Ref, Message: strings.Join(lines, "\n")}
	}
	var de *toml.DecodeError
	if errors.As(err, &de) {
		row, col := de.Position()
		return &ParseError{Source: source.Ref, Message: fmt.Sprintf("%s:%d:%d: %s", source.Ref, row, col, de.Error())}
	}
	return &ParseError{Source: source.Ref, Message: fmt.Sprintf("%s: %v", source.Ref, err)}
}

// acceptedKeys lists the keys valid at the place an unknown key sits. The key
// path says where: choices.<name>.<key> holds a choice, and a block is the
// top of the file or when.<choice>.<answer>.
func acceptedKeys(path []string) []string {
	top := true
	for i := 0; i < len(path); {
		switch {
		case path[i] == "when" && i+3 < len(path):
			top = false
			i += 3
		case path[i] == "choices" && i+3 == len(path):
			return choiceKeys
		default:
			i = len(path)
		}
	}
	if top {
		return topKeys
	}
	return blockKeys
}

// checkRecipe applies the rules TOML typing cannot: a name, defaults that are
// options, and the shape of every use entry, in the recipe and its blocks.
func checkRecipe(d Document) error {
	if strings.TrimSpace(d.Recipe.Name) == "" {
		return &ParseError{Source: d.Source.Ref, Message: d.Source.Ref + ": the recipe has no name; add name = \"...\" at the top"}
	}
	return checkBlock(d, d.Recipe.Block, "")
}

func checkBlock(d Document, b Block, prefix string) error {
	at := func(path string) string { return fmt.Sprintf("%s:%d", d.Source.Ref, d.Line(path)) }
	for _, name := range sortedKeys(b.Choices) {
		c := b.Choices[name]
		if c.Default != "" && len(c.Options) > 0 && !slices.Contains(c.Options, c.Default) {
			return &ParseError{Source: d.Source.Ref, Message: fmt.Sprintf("%s: choice %q has the default %q, which is not one of its options (%s)", at(prefix+"choices."+name), name, c.Default, strings.Join(c.Options, ", "))}
		}
	}
	for _, key := range sortedKeys(b.Use) {
		if _, ok := UseList(b.Use[key]); !ok {
			return &ParseError{Source: d.Source.Ref, Message: fmt.Sprintf("%s: use.%s must be a plugin name, or a list of plugin names", at(prefix+"use."+key), key)}
		}
	}
	for _, choice := range sortedKeys(b.When) {
		for _, answer := range sortedKeys(b.When[choice]) {
			if err := checkBlock(d, b.When[choice][answer], prefix+"when."+choice+"."+answer+"."); err != nil {
				return err
			}
		}
	}
	return nil
}

// UseList reads a use value: one plugin name, or a list of them.
func UseList(v any) ([]string, bool) {
	switch x := v.(type) {
	case string:
		return []string{x}, x != ""
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			s, ok := e.(string)
			if !ok || s == "" {
				return nil, false
			}
			out = append(out, s)
		}
		return out, len(out) > 0
	}
	return nil, false
}

// indexLines walks the file's keys with go-toml's parser and records the line
// of each by its dotted path. A table header and its keys count for every
// path above them, the first line seen winning, so a block's line is where it
// starts and the blocks of a recipe can be put in file order.
func indexLines(data []byte) map[string]int {
	lines := map[string]int{}
	var p unstable.Parser
	p.Reset(data)
	var table []string
	for p.NextExpression() {
		e := p.Expression()
		switch e.Kind {
		case unstable.Table, unstable.ArrayTable:
			parts, line := keyParts(&p, e)
			table = parts
			recordPath(lines, parts, line)
		case unstable.KeyValue:
			parts, line := keyParts(&p, e)
			path := append(slices.Clone(table), parts...)
			recordPath(lines, path, line)
			recordInline(&p, lines, path, e.Value())
		}
	}
	return lines
}

// recordInline records the keys inside an inline table value.
func recordInline(p *unstable.Parser, lines map[string]int, path []string, value *unstable.Node) {
	if value == nil || value.Kind != unstable.InlineTable {
		return
	}
	children := value.Children()
	for children.Next() {
		kv := children.Node()
		parts, line := keyParts(p, kv)
		inner := append(slices.Clone(path), parts...)
		recordPath(lines, inner, line)
		recordInline(p, lines, inner, kv.Value())
	}
}

// keyParts returns the parts of a node's key and the line of the first.
func keyParts(p *unstable.Parser, n *unstable.Node) (parts []string, line int) {
	key := n.Key()
	for key.Next() {
		k := key.Node()
		if line == 0 {
			line = p.Shape(k.Raw).Start.Line
		}
		parts = append(parts, string(k.Data))
	}
	return parts, line
}

func recordPath(lines map[string]int, path []string, line int) {
	for i := 1; i <= len(path); i++ {
		key := strings.Join(path[:i], ".")
		if _, seen := lines[key]; !seen {
			lines[key] = line
		}
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
