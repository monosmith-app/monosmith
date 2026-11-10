package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAGroupThatNamesNoProjectIsWarnedAbout(t *testing.T) {
	lib := &projectgraph.Project{Plugin: "node", Project: protocol.Project{Name: "a", Root: "libs/a", Language: "node", Kind: protocol.ProjectLib}}
	tasks := map[string]workspace.TaskOverride{"build": {}}
	groups := []workspace.Group{
		{Select: []string{"lang:node", "kind:lib"}, Tasks: tasks},
		{Select: []string{"lang:node", "kind:app"}, Tasks: tasks},
	}
	got := GroupFindings([]*projectgraph.Project{lib}, groups)
	if len(got) != 1 || !got[0].Warning || !strings.Contains(got[0].Detail, `"kind:app"`) {
		t.Fatalf("findings = %+v", got)
	}
	if len(GroupFindings([]*projectgraph.Project{lib}, nil)) != 0 {
		t.Error("no groups, nothing to say")
	}
}
