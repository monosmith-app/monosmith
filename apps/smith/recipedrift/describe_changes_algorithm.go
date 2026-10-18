package recipedrift

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// maxDetail is how many changed settings are spelled out in one sentence.
const maxDetail = 3

// levelOf names the level or slot a recipe puts a plugin at ("lint",
// "language"), from the recipe, else from the role it has in a stack. It is
// empty when neither says.
func levelOf(plugin string, res recipe.Resolution, stacks ...protocol.Stack) string {
	for _, u := range res.Uses {
		for _, spec := range u.Plugins {
			if name, _, _ := strings.Cut(spec, "@"); name == plugin {
				return u.Key
			}
		}
	}
	for _, st := range stacks {
		switch {
		case st.Language == plugin:
			return "language"
		case st.Scaffold == plugin:
			return "scaffold"
		case slices.Contains(st.Structure, plugin):
			return "structure"
		case slices.Contains(st.Deploy, plugin):
			return "deploy"
		case slices.Contains(st.Toolchain, plugin):
			return "toolchain"
		}
	}
	return ""
}

func label(plugin, level string) string {
	if level == "" {
		return "the plugin " + plugin
	}
	return "the " + level + " plugin " + plugin
}

// stackChanges says what the recipe's stack changes in the project's: the
// language and scaffold plugin, and the toolchain, structure and deploy
// plugins it adds or drops. added lists the plugins it adds.
func stackChanges(old, now protocol.Stack, res recipe.Resolution) (changes, added []string) {
	if old.Language != now.Language {
		changes = append(changes, fmt.Sprintf("replaces the language plugin %s with %s", old.Language, now.Language))
	}
	if old.Scaffold != now.Scaffold {
		changes = append(changes, fmt.Sprintf("changes the scaffold plugin from %s to %s", orNoneName(old.Scaffold), orNoneName(now.Scaffold)))
	}
	for _, group := range [][2][]string{
		{old.Toolchain, now.Toolchain}, {old.Structure, now.Structure}, {old.Deploy, now.Deploy},
	} {
		for _, p := range group[1] {
			if !slices.Contains(group[0], p) {
				changes = append(changes, "adds "+label(p, levelOf(p, res, now, old)))
				added = append(added, p)
			}
		}
		for _, p := range group[0] {
			if !slices.Contains(group[1], p) {
				changes = append(changes, "drops "+label(p, levelOf(p, res, old, now)))
			}
		}
	}
	return changes, added
}

func orNoneName(name string) string {
	if name == "" {
		return "none"
	}
	return name
}

// settingsChanges says which plugins' settings the recipe now sets
// differently from what smith.toml holds for the project, naming up to three
// settings of each.
func settingsChanges(old, now map[string]map[string]any, res recipe.Resolution, oldStack, newStack protocol.Stack) []string {
	seen := map[string]bool{}
	var plugins []string
	for _, m := range []map[string]map[string]any{old, now} {
		for p := range m {
			if !seen[p] {
				seen[p] = true
				plugins = append(plugins, p)
			}
		}
	}
	sort.Strings(plugins)
	var changes []string
	for _, p := range plugins {
		keys := changedSettings(flatten(old[p]), flatten(now[p]))
		if len(keys) == 0 {
			continue
		}
		detail := keys
		if len(detail) > maxDetail {
			detail = append(append([]string{}, detail[:maxDetail]...), "…")
		}
		changes = append(changes, fmt.Sprintf("changes the settings of %s (%s)", label(p, levelOf(p, res, newStack, oldStack)), strings.Join(detail, ", ")))
	}
	return changes
}

// changedSettings lists, sorted by setting, how now differs from old.
func changedSettings(old, now map[string]string) []string {
	seen := map[string]bool{}
	var names []string
	for _, m := range []map[string]string{old, now} {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				names = append(names, k)
			}
		}
	}
	sort.Strings(names)
	var out []string
	for _, k := range names {
		before, had := old[k]
		after, has := now[k]
		switch {
		case had && has && before == after:
		case had && has:
			out = append(out, fmt.Sprintf("%s: %s → %s", k, before, after))
		case has:
			out = append(out, fmt.Sprintf("%s: set to %s", k, after))
		default:
			out = append(out, k+": removed")
		}
	}
	return out
}

// flatten reduces a settings table to its leaves by dotted path, each value
// as JSON, so a number read as int64 and one read as int compare equal.
func flatten(m map[string]any) map[string]string {
	out := map[string]string{}
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		if table, ok := v.(map[string]any); ok {
			for k, item := range table {
				walk(join(prefix, k), item)
			}
			return
		}
		raw, err := json.Marshal(v)
		if err != nil {
			raw = []byte(fmt.Sprint(v))
		}
		out[prefix] = string(raw)
	}
	for k, v := range m {
		walk(k, v)
	}
	return out
}

func join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

// answerChanges lists the choices the recipe asks that the project has no
// recorded answer for: they take their default, and the upgrade records it.
func answerChanges(recorded, now map[string]string) []string {
	var names []string
	for name := range now {
		if _, ok := recorded[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, fmt.Sprintf("asks a new choice, %s, which takes its default %q", name, now[name]))
	}
	return out
}
