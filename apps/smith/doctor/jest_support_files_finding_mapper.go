package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/legacyjestsupport"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// jestHelperFilesFindings reports the Jest helper files an older smith convert
// wrote at the root (jest.resolver.cjs, jest.asset-transform.cjs): the same
// behaviour is in @monosmith/jest-preset now, so they are clutter a tidy user
// may delete while everything still loads them. Nothing is reported for a
// workspace without them.
func jestHelperFilesFindings(s *session.Session) []protocol.Finding {
	var dirs []string
	for _, p := range s.Graph.Projects {
		dirs = append(dirs, p.Root)
	}
	cleanup := legacyjestsupport.FindCleanup(s.Root, dirs, s.Version)
	if len(cleanup.Files) == 0 {
		return nil
	}
	finding := protocol.Finding{
		Check:   "Jest helper files of an older convert are replaced by " + legacyjestsupport.Package,
		Warning: true,
		Area:    doctorchecks.AreaConfig,
		Problem: strconv.Itoa(len(cleanup.Files)) + " Jest helper file(s) of an older convert are still at the workspace root",
		Detail:  "at the workspace root: " + strings.Join(cleanup.Files, ", ") + ". They stand in for what Nx's Jest plugins provided; " + legacyjestsupport.Package + " does that now, without files to keep",
		Remedy:  "`smith doctor --fix` points the Jest configs at the package, adds it to the root devDependencies and deletes the files",
		Fix:     cleanup.Fix,
	}
	if cleanup.Fix == nil {
		finding.Remedy = cleanup.Manual
	}
	return []protocol.Finding{finding}
}
