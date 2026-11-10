package release

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestPropagationWaitTakesTheProjectsOwnThenTheWorkspaces(t *testing.T) {
	cfg := workspace.Config{Release: workspace.ReleaseSection{PropagationWait: "10m"}}
	own := &projectgraph.Project{Release: &workspace.ProjectRelease{PropagationWait: "0"}}
	plain := &projectgraph.Project{Release: &workspace.ProjectRelease{}}
	for name, c := range map[string]struct {
		p    *projectgraph.Project
		want string
	}{"own": {own, "0"}, "workspace": {plain, "10m"}, "no release table": {&projectgraph.Project{}, "10m"}, "unknown project": {nil, "10m"}} {
		if got := PropagationWait(cfg, c.p); got != c.want {
			t.Errorf("%s: %q, want %q", name, got, c.want)
		}
	}
	if got := PropagationWait(workspace.Config{}, plain); got != "" {
		t.Errorf("nothing set must stay empty (the default is applied where it is parsed), got %q", got)
	}
}

func TestTheReportNamesTheCommandThatMovesAPendingTag(t *testing.T) {
	r := Report{Results: []ReportResult{{Item: Item{Project: "smith", Next: "2.0.0-beta.25"}, Status: "released", PendingTags: []string{"npm dist-tag add monosmith@2.0.0-beta.25 next"}}}}
	for name, text := range map[string]string{"text": r.Text(), "markdown": r.Markdown()} {
		if !strings.Contains(text, "npm dist-tag add monosmith@2.0.0-beta.25 next") || !strings.Contains(text, "pending") {
			t.Errorf("%s:\n%s", name, text)
		}
	}
	if len(Report{}.PendingTags()) != 0 || strings.Contains((Report{}).Text(), "dist-tag") {
		t.Error("a release with nothing pending says nothing about it")
	}
}
