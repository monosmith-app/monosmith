package architecture

import (
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// WorkspaceMembers lists the projects of the workspace at root, by opening it
// as the other commands do. The error says why it cannot be read (not a git
// repository, a plugin that will not start): callers then show the choices
// without saying how they fit.
func WorkspaceMembers(root, version string) ([]Member, error) {
	s, err := session.Open(session.Options{Dir: root, Exact: true, Version: version, Log: io.Discard, AllowCycles: true})
	if err != nil {
		return nil, err
	}
	defer s.Close()
	var members []Member
	for _, p := range s.Graph.Projects {
		members = append(members, Member{Name: p.Name, Language: p.Language})
	}
	return members, nil
}
