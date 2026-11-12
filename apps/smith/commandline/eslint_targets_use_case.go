package commandline

import (
	"io"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/eslintfix"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// eslintTargets lists, for `smith eslint-fix`, the Node projects that have a
// lint task: it opens the workspace that holds dir and reads its graph.
func eslintTargets(version string) eslintfix.Lister {
	return func(dir string) (string, []eslintfix.Target, error) {
		s, err := session.Open(session.Options{Dir: dir, Version: version, Log: io.Discard})
		if err != nil {
			return "", nil, err
		}
		defer s.Close()
		var targets []eslintfix.Target
		for _, name := range s.Graph.Names() {
			p := s.Graph.Get(name)
			if _, lints := p.Tasks["lint"]; lints && p.Language == "node" {
				targets = append(targets, eslintfix.Target{Name: name, Dir: filepath.Join(s.Root, filepath.FromSlash(p.Root))})
			}
		}
		return s.Root, targets, nil
	}
}
