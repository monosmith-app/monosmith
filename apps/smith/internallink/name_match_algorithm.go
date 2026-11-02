package internallink

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// WorkspacePrefix written before a name says "the project of this workspace",
// the prefix pnpm and yarn use for the same meaning.
const WorkspacePrefix = "workspace:"

// StripPrefix removes WorkspacePrefix and reports whether it was there.
func StripPrefix(name string) (string, bool) {
	return strings.CutPrefix(name, WorkspacePrefix)
}

// Match finds the project a name stands for: the project's name, its
// directory, or an identifier it exports (an npm package name, a Python
// distribution name, a module path), in that order. A name that differs only
// in case or in - _ . from an exported identifier matches too, because PyPI
// and npm treat them as one name (shop-core-py is shop_core_py). It returns
// nil when no project matches.
func Match(g *projectgraph.Graph, name string) *projectgraph.Project {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	if p := g.Get(name); p != nil {
		return p
	}
	dir := strings.TrimSuffix(strings.ReplaceAll(name, "\\", "/"), "/")
	dir = strings.TrimPrefix(dir, "./")
	for _, p := range g.Projects {
		if p.Root == dir {
			return p
		}
	}
	for _, p := range g.Projects {
		for _, e := range p.Exports {
			if e == name {
				return p
			}
		}
	}
	key := looseKey(name)
	for _, p := range g.Projects {
		for _, e := range p.Exports {
			if looseKey(e) == key {
				return p
			}
		}
	}
	return nil
}

// looseKey is a name as PyPI compares it: lower case, with runs of - _ . as
// one -.
func looseKey(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(strings.ToLower(name))
}
