package cipipeline

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// packVersionLine is the line smith ci pack prints for a project it packs at
// its planned release version.
func packVersionLine(st release.Stamp) string {
	return fmt.Sprintf("%s: %s -> %s for the pack (the planned release; %s put back after the pack)",
		st.Project, cmp.Or(st.Was, "no version"), st.Version, plural(len(st.Files), "file"))
}

// deployableNames are the projects a deploy may pick up: every app and every
// project with a package task, sorted by name.
func deployableNames(projects []*projectgraph.Project) []string {
	var names []string
	for _, p := range projects {
		if _, packs := p.Tasks[PackTask]; packs || p.Kind == protocol.ProjectApp {
			names = append(names, p.Name)
		}
	}
	slices.Sort(names)
	return names
}
