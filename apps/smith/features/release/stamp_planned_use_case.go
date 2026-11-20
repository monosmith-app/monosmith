package release

import (
	"context"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/features/session"
)

// Stamp is one project that StampPlanned wrote its planned version into.
type Stamp struct {
	// Project is the project's name.
	Project string
	// Was is the version its manifest had before ("" when it had none).
	Was string
	// Version is the planned release version now in its files.
	Version string
	// Files are the files written, workspace-relative.
	Files []string
}

// StampPlanned writes the planned release version into the files of each
// project of projects that the plan releases, the same way a release does
// (the language plugin's version method and the convention files), so what is
// built next carries the version the release will tag. Projects the plan does
// not release keep their manifest version.
//
// Parameters: ctx bounds the plugin calls; s is the open session; plan is the
// release plan (see PlanFor); projects are the project names to stamp; out
// receives a line when a file cannot be put back.
// Returns: the projects stamped, in plan order; a function that puts every
// stamped file back as it was (call it also after a failure; it is safe to
// call when nothing was stamped); and an error when a version could not be
// written, in which case the files already stamped are put back first.
func StampPlanned(ctx context.Context, s *session.Session, plan Plan, projects []string, out io.Writer) ([]Stamp, func(), error) {
	items := PlannedItems(plan, projects)
	if len(items) == 0 {
		return nil, func() {}, nil
	}
	versions, err := SiblingVersions(s.Git, s.Graph, plan)
	if err != nil {
		return nil, func() {}, err
	}
	var stamps []Stamp
	var restores []func()
	restore := func() {
		for i := len(restores) - 1; i >= 0; i-- {
			restores[i]()
		}
	}
	for _, item := range items {
		p := s.Graph.Get(item.Project)
		say := func(format string, args ...any) { fmt.Fprintf(out, item.Project+": "+format+"\n", args...) }
		var res ReportResult
		files, undo, err := stampVersion(ctx, s, p, item, versions, Options{Out: out}, &res, say)
		if err != nil {
			restore()
			return nil, func() {}, fmt.Errorf("%s: %w", item.Project, err)
		}
		restores = append(restores, undo)
		if len(files) > 0 {
			stamps = append(stamps, Stamp{Project: item.Project, Was: p.Version, Version: item.Next, Files: files})
		}
	}
	return stamps, restore, nil
}
