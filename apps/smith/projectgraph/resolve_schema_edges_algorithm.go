package projectgraph

import (
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ResolveSchemaEdges links the owner of a shared schema to its consumers.
// files are the workspace's files, workspace-relative.
//
// A project owns the schema files (OpenAPI, AsyncAPI, Protocol Buffers,
// GraphQL) inside its root, plus the files its [projects.<name>] schemas globs
// match, wherever they are. It consumes the files its consumes_schemas globs
// match. A consumer then depends on the owner of every schema it consumes (a
// file no project owns by that name falls back to the project whose root holds
// it), so a schema change affects the clients generated from it. Run it after
// ResolveEdges: it adds to Deps.
func ResolveSchemaEdges(projects []*Project, files []string, cfg workspace.Config) {
	byRoot := func(file string) *Project {
		var best *Project
		for _, p := range projects {
			if file == p.Root || strings.HasPrefix(file, p.Root+"/") {
				if best == nil || len(p.Root) > len(best.Root) {
					best = p
				}
			}
		}
		return best
	}
	owners := map[string]string{}
	for _, f := range files {
		if SchemaKind(f) == "" {
			continue
		}
		if p := byRoot(f); p != nil {
			owners[f] = p.Name
		}
	}
	for _, p := range projects {
		for _, glob := range cfg.Projects[p.Name].Schemas {
			for _, f := range files {
				if ok, _ := doublestar.Match(glob, f); ok {
					owners[f] = p.Name
				}
			}
		}
	}
	for _, p := range projects {
		owned := map[string]bool{}
		for f, owner := range owners {
			if owner == p.Name {
				owned[f] = true
			}
		}
		p.Schemas = sortedKeys(owned)
	}
	for _, p := range projects {
		deps := map[string]bool{}
		consumed := map[string]bool{}
		for _, glob := range cfg.Projects[p.Name].ConsumesSchemas {
			for _, f := range files {
				if ok, _ := doublestar.Match(glob, f); !ok {
					continue
				}
				consumed[f] = true
				owner, known := owners[f]
				if !known {
					if r := byRoot(f); r != nil {
						owner = r.Name
					}
				}
				if owner != "" && owner != p.Name {
					deps[owner] = true
				}
			}
		}
		p.Consumes = sortedKeys(consumed)
		p.SchemaDeps = sortedKeys(deps)
		for _, d := range p.Deps {
			deps[d] = true
		}
		p.Deps = sortedKeys(deps)
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
