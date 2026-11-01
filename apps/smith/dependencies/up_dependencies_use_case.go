package dependencies

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// UpReport is `smith up --json`.
type UpReport struct {
	Updates []Update `json:"updates"`
	// Unreachable are names a registry could not be asked about, by plugin.
	Unreachable map[string]map[string]string `json:"unreachable,omitempty"`
	// Tools are the tools each plugin pins against their newest releases.
	Tools []Tool `json:"tools,omitempty"`
	// Files are the manifests (and lockfiles) up changed.
	Files []string `json:"files,omitempty"`
}

// Tool is one pinned tool, with the plugin that pins it.
type Tool struct {
	Plugin string `json:"plugin"`
	protocol.ToolPin
}

// FindAvailable asks every plugin that can for the newest releases of what
// its projects declare, and of the tools it pins.
func FindAvailable(s *session.Session) (UpReport, error) {
	report := UpReport{Unreachable: map[string]map[string]string{}}
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		projects, local := servedBy(s, name)
		if len(projects) == 0 || !c.Supports(protocol.MethodLatest) {
			continue
		}
		var declared protocol.DeclaredResult
		if c.Supports(protocol.MethodDeclared) {
			if err := c.Call(protocol.MethodDeclared, protocol.DeclaredParams{Projects: projects}, &declared, timeout); err != nil {
				return report, fmt.Errorf("plugin %s: %w", name, err)
			}
		}
		seen := map[string]bool{}
		var names []string
		for _, d := range declared.Declarations {
			if d.Rewritable && !d.Peer && !local[d.Name] && !seen[d.Name] {
				seen[d.Name] = true
				names = append(names, d.Name)
			}
		}
		sort.Strings(names)
		var latest protocol.LatestResult
		if err := c.Call(protocol.MethodLatest, protocol.LatestParams{Names: names, Tools: true}, &latest, timeout); err != nil {
			return report, fmt.Errorf("plugin %s: %w", name, err)
		}
		report.Updates = append(report.Updates, FindUpdates(name, declared, latest.Latest, local)...)
		if len(latest.Unreachable) > 0 {
			report.Unreachable[name] = latest.Unreachable
		}
		for _, t := range latest.Tools {
			report.Tools = append(report.Tools, Tool{Plugin: name, ToolPin: t})
		}
	}
	sort.SliceStable(report.Updates, func(i, j int) bool {
		return levelIndex(report.Updates[i].Level) < levelIndex(report.Updates[j].Level)
	})
	MarkPresets(report.Updates, s.Config.Dependencies.Presets)
	return report, nil
}

// Apply rewrites the chosen updates through their plugins and installs, so
// lockfiles and the venv follow, unless install is false.
func Apply(s *session.Session, updates []Update, install bool) ([]string, error) {
	byPlugin := map[string][]Update{}
	var plugins []string
	for _, u := range updates {
		if _, ok := byPlugin[u.Plugin]; !ok {
			plugins = append(plugins, u.Plugin)
		}
		byPlugin[u.Plugin] = append(byPlugin[u.Plugin], u)
	}
	sort.Strings(plugins)
	var files []string
	for _, name := range plugins {
		c := s.Plugins.Get(name)
		projects, _ := servedBy(s, name)
		changed, err := rewriteAndInstall(c, rewrites(byPlugin[name]), projects, install)
		files = append(files, changed...)
		if err != nil {
			return files, fmt.Errorf("plugin %s: %w", name, err)
		}
	}
	return files, nil
}

// Choose keeps the updates of the packages named.
func Choose(updates []Update, names []string) (chosen []Update, unknown []string) {
	for _, n := range names {
		found := false
		for _, u := range updates {
			if u.Name == n {
				chosen, found = append(chosen, u), true
			}
		}
		if !found {
			unknown = append(unknown, n)
		}
	}
	return chosen, unknown
}

// PrintUp writes the report for people: updates grouped by level, what could
// not be asked, and the pinned tools.
func PrintUp(w io.Writer, r UpReport) {
	if len(r.Updates) == 0 {
		fmt.Fprintln(w, "✔ Every declared dependency is at its latest release.")
	}
	current := ""
	for _, u := range r.Updates {
		if u.Group() != current {
			current = u.Group()
			if current == "preset" {
				fmt.Fprintln(w, "preset (your team's shared config, bumped in every project that uses it)")
			} else {
				fmt.Fprintln(w, current)
			}
		}
		fmt.Fprintf(w, "  %-28s %s → %s  %s\n", u.Name, u.Current, u.Latest, strings.Join(u.Projects(), ", "))
	}
	for plugin, names := range r.Unreachable {
		for name, why := range names {
			fmt.Fprintf(w, "⊘ %s (%s): the registry could not be asked (%s)\n", name, plugin, why)
		}
	}
	if len(r.Tools) > 0 {
		fmt.Fprintln(w, "\nTools smith pins (informational):")
		for _, t := range r.Tools {
			switch {
			case t.Latest == "":
				fmt.Fprintf(w, "  %-16s %s  (its release feed could not be reached)\n", t.Name, t.Pinned)
			case t.Newer:
				fmt.Fprintf(w, "  %-16s %s → %s available\n", t.Name, t.Pinned, t.Latest)
			default:
				fmt.Fprintf(w, "  %-16s %s ✔ (latest %s)\n", t.Name, t.Pinned, t.Latest)
			}
		}
	}
}
