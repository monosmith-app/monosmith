package dependencies

import (
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// isDev reports whether a declaration is a dev tool: devDependencies,
// Python's dev extra, Cargo's dev-dependencies.
func isDev(d protocol.Declaration) bool {
	return strings.HasPrefix(strings.ToLower(d.Section), "dev")
}

// FindFragmentation finds the production dependencies one plugin's projects
// declare at more than one spec. Dev tools are the other check
// (FindMisplacedDevTools); peers, the workspace's own packages and the
// dependencies allowed to differ do not count.
func FindFragmentation(plugin string, declared protocol.DeclaredResult, local map[string]bool, allow []string) []Drift {
	production := protocol.DeclaredResult{Installed: declared.Installed}
	for _, d := range declared.Declarations {
		if !isDev(d) && !slices.Contains(allow, d.Name) {
			production.Declarations = append(production.Declarations, d)
		}
	}
	return FindDrift(plugin, production, local)
}

// MisplacedDevTool is a dev tool declared in more than one manifest, at
// least one of them a project's: dev tools live at the root, once.
type MisplacedDevTool struct {
	Plugin string `json:"plugin"`
	Name   string `json:"name"`
	// Sites are the manifests declaring it, by project ("(root)" for the
	// root manifest).
	Sites []string `json:"sites"`
}

// FindMisplacedDevTools finds the dev tools one plugin's manifests declare
// more than once with a project among the declarers. Peers, the workspace's
// own packages and the allowed dependencies do not count.
func FindMisplacedDevTools(plugin string, declared protocol.DeclaredResult, local map[string]bool, allow []string) []MisplacedDevTool {
	byName := map[string][]protocol.Declaration{}
	for _, d := range declared.Declarations {
		if isDev(d) && !d.Peer && !local[d.Name] && !slices.Contains(allow, d.Name) {
			byName[d.Name] = append(byName[d.Name], d)
		}
	}
	var out []MisplacedDevTool
	for name, sites := range byName {
		inProject := false
		var places []string
		for _, s := range sites {
			inProject = inProject || s.Project != ""
			places = append(places, where(s))
		}
		if len(sites) < 2 || !inProject {
			continue
		}
		sort.Strings(places)
		out = append(out, MisplacedDevTool{Plugin: plugin, Name: name, Sites: places})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
