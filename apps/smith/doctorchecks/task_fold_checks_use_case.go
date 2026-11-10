package doctorchecks

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskFoldChecks suggests [[groups]] when two or more projects repeat the
// same task settings in smith.toml, and offers the rewrite as a fix. It is a
// warning, and says nothing when nothing repeats.
func TaskFoldChecks(s *session.Session) []protocol.Finding {
	raw, err := os.ReadFile(filepath.Join(s.Root, workspace.FileName))
	if err != nil {
		return nil
	}
	if f, ok := TaskFoldFinding(s.Graph.Projects, s.Config, string(raw)); ok {
		return []protocol.Finding{f}
	}
	return nil
}
