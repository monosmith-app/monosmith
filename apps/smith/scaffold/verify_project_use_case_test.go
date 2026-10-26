package scaffold

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestVerifyTasksAreTheBuildAndTestTheProjectHas(t *testing.T) {
	task := func(names ...string) *projectgraph.Project {
		p := &projectgraph.Project{Project: protocol.Project{Name: "x", Tasks: map[string]protocol.Task{}}}
		for _, n := range names {
			p.Tasks[n] = protocol.Task{}
		}
		return p
	}
	for _, c := range []struct {
		project *projectgraph.Project
		want    []string
	}{
		{task("lint", "test", "build"), []string{"build", "test"}},
		{task("test"), []string{"test"}},
		{task("build", "lint"), []string{"build"}},
		{task("lint"), nil},
		{nil, nil},
	} {
		if got := verifyTasks(c.project); !slices.Equal(got, c.want) {
			t.Errorf("%v: got %v, want %v", c.project, got, c.want)
		}
	}
}
