package recipestack

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Declaration is what a recipe adds to smith.toml for one new project.
type Declaration struct {
	// Project is the new project's name.
	Project string
	// Stack is declared under [stacks.<name>] unless Reuse says smith.toml
	// or a plugin already has it.
	Stack protocol.Stack
	Reuse bool
	// Settings are the recipe's settings by plugin, for [projects.<name>.settings].
	Settings map[string]map[string]any
	// Recipe is recorded in [projects.<name>.recipe].
	Recipe workspace.ProjectRecipe
}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func key(k string) string {
	if bareKey.MatchString(k) {
		return k
	}
	return configedit.Quote(k)
}

// Render is the smith.toml text of a declaration: a comment saying where it
// came from, the stack, the project's recipe record and its settings.
func Render(d Declaration) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Written by smith from the recipe %s for the project %s.\n", d.Recipe.Name, d.Project)
	if !d.Reuse {
		fmt.Fprintf(&b, "[stacks.%s]\n", key(d.Stack.Name))
		fmt.Fprintf(&b, "description = %s\n", configedit.Quote(d.Stack.Description))
		fmt.Fprintf(&b, "language = %s\n", configedit.Quote(d.Stack.Language))
		list(&b, "toolchain", d.Stack.Toolchain)
		list(&b, "structure", d.Stack.Structure)
		if d.Stack.Scaffold != "" {
			fmt.Fprintf(&b, "scaffold = %s\n", configedit.Quote(d.Stack.Scaffold))
		}
		list(&b, "deploy", d.Stack.Deploy)
		b.WriteString("\n")
	}
	base := "projects." + key(d.Project)
	fmt.Fprintf(&b, "[%s.recipe]\nname = %s\n", base, configedit.Quote(d.Recipe.Name))
	if d.Recipe.Version != "" {
		fmt.Fprintf(&b, "version = %s\n", configedit.Quote(d.Recipe.Version))
	}
	fmt.Fprintf(&b, "source = %s\n", configedit.Quote(d.Recipe.Source))
	if len(d.Recipe.Answers) > 0 {
		fmt.Fprintf(&b, "\n[%s.recipe.answers]\n", base)
		names := make([]string, 0, len(d.Recipe.Answers))
		for name := range d.Recipe.Answers {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(&b, "%s = %s\n", key(name), configedit.Quote(d.Recipe.Answers[name]))
		}
	}
	plugins := make([]string, 0, len(d.Settings))
	for p := range d.Settings {
		plugins = append(plugins, p)
	}
	sort.Strings(plugins)
	for _, p := range plugins {
		table(&b, base+".settings."+key(p), d.Settings[p])
	}
	return b.String()
}

func list(b *strings.Builder, name string, items []string) {
	if len(items) == 0 {
		return
	}
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = configedit.Quote(item)
	}
	fmt.Fprintf(b, "%s = [%s]\n", name, strings.Join(quoted, ", "))
}

// table writes a settings table: its plain values first, then each nested
// table under its own header.
func table(b *strings.Builder, header string, values map[string]any) {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprintf(b, "\n[%s]\n", header)
	var nested []string
	for _, name := range names {
		if _, isTable := values[name].(map[string]any); isTable {
			nested = append(nested, name)
			continue
		}
		text, err := value(values[name])
		if err != nil {
			fmt.Fprintf(b, "# %s: cannot be written (%v)\n", key(name), err)
			continue
		}
		fmt.Fprintf(b, "%s = %s\n", key(name), text)
	}
	for _, name := range nested {
		table(b, header+"."+key(name), values[name].(map[string]any))
	}
}

// value is a setting's value as TOML. Strings are basic strings, as in the
// rest of smith.toml, and so are those in a list; go-toml writes any other
// value (numbers, booleans, lists of them, inline values).
func value(v any) (string, error) {
	switch v := v.(type) {
	case string:
		return configedit.Quote(v), nil
	case []string:
		items := make([]any, len(v))
		for i, s := range v {
			items[i] = s
		}
		return value(items)
	case []any:
		parts := make([]string, len(v))
		for i, item := range v {
			if _, isTable := item.(map[string]any); isTable {
				return inline(v)
			}
			text, err := value(item)
			if err != nil {
				return "", err
			}
			parts[i] = text
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	}
	return inline(v)
}

// inline lets go-toml write a value after its key and strips the key again.
func inline(v any) (string, error) {
	raw, err := toml.Marshal(map[string]any{"v": v})
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(raw))
	text = strings.TrimPrefix(text, "v = ")
	if strings.Contains(text, "\n") {
		return "", fmt.Errorf("a value of more than one line")
	}
	return text, nil
}
