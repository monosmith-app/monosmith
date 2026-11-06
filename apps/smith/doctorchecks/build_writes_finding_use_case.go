package doctorchecks

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// BuildWritesFinding fails when a build script writes directories its build
// task does not declare as outputs: a cache hit restores only the declared
// ones, so the real folder stays empty and whatever reads it (a dependent's
// types, the package's own scripts) breaks on a warm cache. The fix declares
// the outputs in smith.toml, which only changes what the cache restores.
//
// Parameters: s is the open session. Returns the finding, with a fix when
// smith.toml can be read.
func BuildWritesFinding(s *session.Session) protocol.Finding {
	gaps := buildoutputs.FindGaps(s.Root, s.Graph.Projects)
	raw, err := os.ReadFile(filepath.Join(s.Root, workspace.FileName))
	if err != nil {
		return buildWritesFromGaps(gaps, "", false)
	}
	return buildWritesFromGaps(gaps, string(raw), true)
}
