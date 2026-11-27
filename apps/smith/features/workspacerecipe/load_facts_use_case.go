package workspacerecipe

import (
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/features/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/features/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Load opens the workspace at root and gathers its Facts: each project's
// language plugin and stack, the code layout smith.toml turns on, and the
// tasks each toolchain plugin of those stacks provides. version is smith's,
// which the built-in plugins report as their own.
//
// It returns an error when the workspace cannot be opened or a stack it
// names is unknown.
func Load(root, version string) (Facts, error) {
	s, err := session.Open(session.Options{Dir: root, Exact: true, Version: version, Log: io.Discard, AllowCycles: true})
	if err != nil {
		return Facts{}, err
	}
	defer s.Close()
	var f Facts
	toolchain := map[string]bool{}
	for _, p := range s.Graph.Projects {
		fact := ProjectFact{Language: p.Plugin, StackName: p.Stack}
		if p.Stack != "" && p.Stack != protocol.NoStack {
			st, err := s.Stacks.Get(p.Stack)
			if err != nil {
				return Facts{}, err
			}
			fact.Stack = st
			for _, name := range st.Toolchain {
				toolchain[name] = true
			}
		}
		f.Projects = append(f.Projects, fact)
	}
	if a, err := architecture.Current(s.Root); err == nil && a.Name != architecture.None {
		f.Layouts = append(f.Layouts, Layout{Name: a.Name, Languages: a.AppliesTo()})
	}
	lookup := recipe.Catalog{Root: s.Root, Version: version, External: s.Config.Plugins.External}.Lookup
	f.Provides = map[string][]string{}
	for name := range toolchain {
		if info, err := lookup(name); err == nil {
			f.Provides[name] = info.Provides
		}
	}
	return f, nil
}
