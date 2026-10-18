package recipechoice

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// Describe says what a resolved recipe will use for each level or slot, with
// the recipe line that chose it, and the settings it carries. It is shown
// before anything is written.
func Describe(res recipe.Resolution) string {
	var b strings.Builder
	title := "Recipe " + res.Name
	if res.Version != "" {
		title += " " + res.Version
	}
	if answers := answersOf(res); answers != "" {
		title += " (" + answers + ")"
	}
	b.WriteString(title + " will use:\n")
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	for _, u := range res.Uses {
		fmt.Fprintf(w, "  %s\t%s\t%s\n", u.Key, strings.Join(u.Plugins, ", "), u.Origin)
	}
	_ = w.Flush()
	for _, line := range settingLines(res) {
		b.WriteString("  " + line + "\n")
	}
	return b.String()
}

// answersOf lists the answers in the order the choices were asked.
func answersOf(res recipe.Resolution) string {
	var parts []string
	for _, q := range res.Asked {
		if q.Answer != "" {
			parts = append(parts, q.Name+"="+q.Answer)
		}
	}
	return strings.Join(parts, ", ")
}

// settingLines lists the recipe's settings as plugin.path = value.
func settingLines(res recipe.Resolution) []string {
	var lines []string
	var walk func(prefix string, table map[string]any)
	walk = func(prefix string, table map[string]any) {
		keys := make([]string, 0, len(table))
		for k := range table {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if sub, ok := table[k].(map[string]any); ok {
				walk(prefix+k+".", sub)
				continue
			}
			lines = append(lines, fmt.Sprintf("setting %s%s = %v", prefix, k, table[k]))
		}
	}
	plugins := make([]string, 0, len(res.Settings))
	for p := range res.Settings {
		plugins = append(plugins, p)
	}
	sort.Strings(plugins)
	for _, p := range plugins {
		if table, ok := res.Settings[p].(map[string]any); ok {
			walk(p+".", table)
		}
	}
	return lines
}
