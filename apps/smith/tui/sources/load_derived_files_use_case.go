package sources

import (
	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// loadDerivedFiles reads the derived files of every project, the way
// `smith derive --check` finds them, and writes nothing. It returns nil when
// the workspace did not turn derived files on, and also when they cannot be
// read: Doctor reports that problem with its remedy, and the project list
// must still load.
//
// s is the open session.
func loadDerivedFiles(s *session.Session) map[string][]projectlist.DerivedFile {
	if !s.Config.Derive.Enabled {
		return nil
	}
	plan, _, err := derive.BuildPlan(s)
	if err != nil {
		return nil
	}
	return derivedByProject(plan)
}

// derivedByProject groups the files of a derive plan by project, with the
// files left to the project (they differ from its siblings) after those smith
// derives.
func derivedByProject(plan derive.Plan) map[string][]projectlist.DerivedFile {
	out := map[string][]projectlist.DerivedFile{}
	for _, i := range plan.Items {
		out[i.File.Project] = append(out[i.File.Project], projectlist.DerivedFile{
			Path: i.File.Path, Renderer: i.File.Renderer, Status: string(i.Status), Detail: i.Detail, Keys: i.File.Mode == protocol.DeriveKeys,
		})
	}
	for _, d := range plan.Diverging {
		out[d.Project] = append(out[d.Project], projectlist.DerivedFile{
			Path: d.Path, Renderer: d.Renderer, Status: projectlist.StatusDiffers, Detail: d.Detail,
		})
	}
	return out
}
