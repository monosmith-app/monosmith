package commandline

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/features/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
)

// projectFacts is how `smith config` learns what a project's plugin reports
// about it (its folder, language, kind and version): it opens the workspace
// with its plugins, as every command that names a project does.
//
// Parameters: version is the running smith's version.
// Returns: the loader configedit.Command takes.
func projectFacts(version string) configedit.ProjectLoader {
	return func(dir, name string, log io.Writer) (configedit.ProjectFacts, error) {
		s, err := session.Open(session.Options{Dir: dir, Version: version, Log: log})
		if err != nil {
			return configedit.ProjectFacts{}, err
		}
		defer s.Close()
		p := s.Graph.Get(name)
		if p == nil {
			return configedit.ProjectFacts{}, fmt.Errorf("%s is not a project of this workspace\n→ smith projects lists them", name)
		}
		return configedit.ProjectFacts{Name: p.Name, Root: p.Root, Language: p.Language, Kind: string(p.Kind), Version: p.Version}, nil
	}
}
