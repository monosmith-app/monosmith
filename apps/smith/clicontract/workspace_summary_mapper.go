package clicontract

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// WorkspaceSummary describes a workspace in a few lines a newcomer can read:
// its name, projects by kind and language, the stacks and CI in use, and
// where releases come from, and how full the local cache is. smith config list
// has the rest.
func WorkspaceSummary(root string, cfg workspace.Config, graph *projectgraph.Graph, cache *CacheInfo) string {
	var b strings.Builder
	row := func(label, value string) {
		fmt.Fprintf(&b, "  %-9s %s\n", label, value)
	}
	name := cfg.Workspace.Name
	if name == "" {
		name = "(unnamed)"
	}
	fmt.Fprintf(&b, "workspace %s (%s)\n", name, root)
	kinds, languages, stacks := map[string]int{}, map[string]int{}, map[string]int{}
	for _, p := range graph.Projects {
		kind := string(p.Kind)
		if kind == "" {
			kind = "project"
		}
		kinds[kind]++
		languages[p.Plugin]++
		if p.Stack != "" {
			stacks[p.Stack]++
		}
	}
	row("projects", fmt.Sprintf("%d%s", len(graph.Projects), counts(kinds, true)))
	if len(languages) > 0 {
		row("languages", strings.TrimPrefix(counts(languages, false), " "))
	}
	stackText := "none recorded"
	if len(stacks) > 0 {
		stackText = strings.TrimPrefix(counts(stacks, false), " ")
	}
	if len(cfg.DefaultStacks) > 0 {
		defaults := make([]string, 0, len(cfg.DefaultStacks))
		for language, stack := range cfg.DefaultStacks {
			defaults = append(defaults, stack+" for "+language)
		}
		sort.Strings(defaults)
		stackText += "; default " + strings.Join(defaults, ", ")
	}
	row("stacks", stackText)
	providers := "none"
	if len(cfg.CI.Providers) > 0 {
		providers = strings.Join(cfg.CI.Providers, ", ")
	}
	row("CI", providers)
	branch := cfg.Release.Branch
	if branch == "" {
		branch = "main"
	}
	row("releases", "from "+branch)
	if cache != nil {
		row("cache", cache.Line())
	}
	b.WriteString("\nsmith config list shows every setting; smith projects lists the projects.\n")
	return b.String()
}

// counts renders {"app": 1, "package": 22} as " (1 app, 22 package)" with
// parentheses, or "1 app, 22 package" without.
func counts(m map[string]int, parens bool) string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		label := name
		if parens && m[name] != 1 {
			label += "s"
		}
		parts[i] = fmt.Sprintf("%d %s", m[name], label)
	}
	if len(parts) == 0 {
		return ""
	}
	if parens {
		return " (" + strings.Join(parts, ", ") + ")"
	}
	return " " + strings.Join(parts, ", ")
}
