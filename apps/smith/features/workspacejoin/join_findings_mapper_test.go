package workspacejoin

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFindingsCarryThePluginsFixAndTheWayToStayOut(t *testing.T) {
	fix := &protocol.Fix{Description: `add "apps/*" to the root package.json "workspaces", then run npm install`, Safe: true}
	got := Findings(
		[]protocol.Project{{Name: "web", Root: "packages/web"}, {Name: "api", Root: "apps/api"}},
		protocol.JoinResult{Projects: []protocol.JoinState{
			{Project: "web", Joined: true, Workspace: "w"},
			{Project: "api", Workspace: "w", Problem: "is not in w: npm install does not reach it", Fix: fix},
			{Project: "ghost", Workspace: "w"},
		}},
	)
	if len(got) != 2 {
		t.Fatalf("a project the plugin was not asked about must be dropped: %+v", got)
	}
	if !got[0].OK || got[0].Check != "packages/web is joined to its language's workspace" || !IsJoinCheck(got[0].Check) {
		t.Errorf("passed: %+v", got[0])
	}
	f := got[1]
	if f.OK || f.Project != "api" || f.Fix != fix || !strings.HasPrefix(f.Remedy, "smith doctor --fix: add \"apps/*\"") || !strings.Contains(f.Remedy, `exclude = ["apps/api"]`) {
		t.Errorf("failed: %+v", f)
	}
}

func TestFindingsWithoutAFixSayHowToJoinByHand(t *testing.T) {
	got := Findings([]protocol.Project{{Name: "a", Root: "apps/a"}}, protocol.JoinResult{Projects: []protocol.JoinState{{Project: "a", Workspace: "the solution X.sln"}}})
	if len(got) != 1 || got[0].Problem != "is not joined to the solution X.sln" || !strings.HasPrefix(got[0].Remedy, "join it by hand") {
		t.Fatalf("%+v", got)
	}
}

func TestUnreachedLineNamesTheProjectsAndTheStep(t *testing.T) {
	if UnreachedLine(nil) != "" {
		t.Error("nothing unreached says nothing")
	}
	line := UnreachedLine([]string{"faviconit-api"})
	if !strings.Contains(line, "not installed: faviconit-api") || !strings.Contains(line, "→ smith doctor --fix") {
		t.Errorf("%q", line)
	}
}
