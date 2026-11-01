package agentdoc

import (
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
)

// doneTasks are the tasks that mean done, in the order they run.
var doneTasks = []string{"lint", "typecheck", "test", "build"}

// Collect reads the rules of the open workspace: the projects and their
// conventions from the graph, the stacks and the release branch from
// smith.toml, and the derived files from smith.lock. It returns an error when
// smith.lock cannot be read.
func Collect(s *session.Session) (Facts, error) {
	rec, err := derivelock.Read(s.Root)
	if err != nil {
		return Facts{}, err
	}
	f := Facts{Name: s.Config.Workspace.Name, Derive: s.Config.Derive.Enabled, Branch: s.Config.Release.Branch}
	if f.Branch == "" {
		f.Branch = "main"
	}
	arch := map[string]bool{}
	langs := map[string]bool{}
	tasks := map[string]bool{}
	for _, p := range s.Graph.Projects {
		pr := Project{Name: p.Name, Root: p.Root, Kind: string(p.Kind), Language: p.Language, Stack: p.Stack}
		if st, ok := s.Config.Stacks[p.Stack]; ok {
			pr.Choices = st.Choices
		}
		f.Projects = append(f.Projects, pr)
		langs[p.Language] = true
		for _, c := range p.Conventions {
			if c != architecture.None && slices.Contains(architecture.Names(), c) {
				arch[c] = true
			}
		}
		for t := range p.Tasks {
			tasks[t] = true
		}
	}
	sort.Slice(f.Projects, func(i, j int) bool { return f.Projects[i].Root < f.Projects[j].Root })
	for _, a := range architecture.Catalog() {
		if arch[a.Name] {
			arch := Architecture{Name: a.Name, Meaning: a.Meaning}
			if rs, ok := rules.For(a.Name); ok {
				arch.Rules = rs.Summary
				for _, l := range rs.Layers {
					arch.Layers = append(arch.Layers, Layer{Name: l.Name, Meaning: l.Meaning, Folders: l.Folders})
				}
			}
			f.Architectures = append(f.Architectures, arch)
		}
	}
	for l := range langs {
		f.Languages = append(f.Languages, l)
	}
	sort.Strings(f.Languages)
	for _, t := range doneTasks {
		if tasks[t] {
			f.Tasks = append(f.Tasks, t)
		}
	}
	f.Derived = rec.OwnedBySmith()
	for _, e := range rec.Files {
		if e.Owner == derivelock.OwnerYou {
			f.Ejected = append(f.Ejected, e.Path)
		}
	}
	sort.Strings(f.Ejected)
	f.Ejected = slices.Compact(f.Ejected)
	return f, nil
}
