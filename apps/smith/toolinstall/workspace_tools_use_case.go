package toolinstall

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

// coreTools are the tools smith itself needs: git always, the GitHub CLI when
// the workspace keeps its pipelines on GitHub.
func coreTools(root string) []protocol.ToolNeed {
	core := toolcatalog.For("git")
	if info, err := os.Stat(filepath.Join(root, ".github")); err == nil && info.IsDir() {
		core = append(core, toolcatalog.For("gh")...)
	}
	return core
}

// sourcesOf reads what the loaded plugins declare.
func sourcesOf(s *session.Session) []Source {
	var out []Source
	if s.Plugins == nil {
		return nil
	}
	for _, c := range s.Plugins.Clients {
		if tools := ForSystem(c.Info.Tools, runtime.GOOS); len(tools) > 0 {
			out = append(out, Source{Plugin: c.Info.Plugin.Name, Languages: c.Info.Languages, Tools: tools})
		}
	}
	return out
}

// usesOf lists the projects of the workspace with the choices of their stacks.
func usesOf(s *session.Session) []Use {
	var out []Use
	if s.Graph == nil {
		return nil
	}
	for _, p := range s.Graph.Projects {
		u := Use{Project: p.Name, Language: p.Language, Root: filepath.Join(s.Root, filepath.FromSlash(p.Root))}
		if p.Stack != "" {
			if st, err := s.Stacks.Get(p.Stack); err == nil {
				u.Choices = append(slices.Clone(st.Choices), st.Toolchain...)
			}
		}
		out = append(out, u)
	}
	return out
}

// WorkspaceNeeds lists the tools an open workspace needs.
//
// Parameters: s is the open session; extraLanguages and extraChoices add the
// tools of a language or a choice that no project uses yet (smith add --lang).
func WorkspaceNeeds(s *session.Session, extraLanguages, extraChoices []string) []Need {
	return Needed(s.Root, coreTools(s.Root), sourcesOf(s), usesOf(s), extraLanguages, extraChoices)
}

// Declares reports whether smith tools install knows a tool by that name.
//
// Parameters: s is the open session; name is the tool's name.
func Declares(s *session.Session, name string) bool {
	_, unknown := NamedNeeds(s, []string{name})
	return len(unknown) == 0
}

// NamedNeeds finds tools by name among everything the plugins declare.
//
// Returns: the tools in the order asked, and the names nothing declares.
func NamedNeeds(s *session.Session, names []string) ([]Need, []string) {
	return Lookup(s.Root, coreTools(s.Root), sourcesOf(s), names)
}
