package tui

import "github.com/monosmith-app/monosmith/apps/smith/features/tui/projectlist"

// pendingNext is the version the pending release makes of a project, and the
// version the release starts from. next is "" when the Release tab has no plan
// that releases the project. The data is the plan of the Release tab, so the
// two tabs never disagree.
func (s Shell) pendingNext(project string) (current, next string) {
	if !s.release.Loaded || s.release.Err != nil {
		return "", ""
	}
	for _, it := range s.release.Plan.Items {
		if it.Project == project {
			return it.Current, it.Next
		}
	}
	return "", ""
}

// versionOf is the text of a project's version in w cells: "current → next"
// while a release is pending, else the current version.
func (s Shell) versionOf(p projectlist.Project, w int) string {
	current, next := s.pendingNext(p.Name)
	if p.Version != "" {
		current = p.Version
	}
	return projectlist.VersionText(current, next, w)
}
