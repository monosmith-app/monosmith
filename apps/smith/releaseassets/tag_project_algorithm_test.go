package releaseassets

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestProjectOfTagPicksTheBuildProjectWithTheLongestPrefix(t *testing.T) {
	build := &workspace.ProjectRelease{Build: "make", Assets: []string{"out/*"}}
	mvd := project("mvd", false, build)
	tui := project("mvd-tui", false, build)
	plain := project("lib", false, &workspace.ProjectRelease{})
	bare := project("bare", true, build)
	projects := []*projectgraph.Project{plain, bare, mvd, tui}
	for _, tc := range []struct{ tag, name, version string }{
		{"mvd@0.11.1", "mvd", "0.11.1"},
		{"mvd-tui@0.8.1", "mvd-tui", "0.8.1"},
		{"v1.2.0", "bare", "v1.2.0"},
		{"lib@1.0.0", "bare", "lib@1.0.0"},
	} {
		p, version := projectOfTag(projects, tc.tag)
		if p == nil || p.Name != tc.name || version != tc.version {
			t.Errorf("%s: got %v %q, want %s %q", tc.tag, p, version, tc.name, tc.version)
		}
	}
	if p, _ := projectOfTag([]*projectgraph.Project{plain, mvd}, "lib@1.0.0"); p != nil {
		t.Errorf("a project without a build command must not match: %v", p)
	}
	if p, _ := projectOfTag([]*projectgraph.Project{mvd}, "mvd@"); p != nil {
		t.Errorf("a tag without a version must not match: %v", p)
	}
}

func project(name string, bare bool, r *workspace.ProjectRelease) *projectgraph.Project {
	p := &projectgraph.Project{BareTags: bare, Release: r}
	p.Name = name
	return p
}
