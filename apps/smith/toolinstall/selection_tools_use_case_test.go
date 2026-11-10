package toolinstall

import (
	"context"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSelectionStatusesListsLanguageHostAndShellToolsWithTheirState(t *testing.T) {
	sources := []Source{
		{Plugin: "node", Languages: []string{"node"}, Tools: []protocol.ToolNeed{
			{Name: "node"},
			{Name: "aws-sam-cli", Choice: "aws-lambda"},
			{Name: "azure-functions-core-tools", Choice: "azure-functions"},
		}},
		{Plugin: "tauri", Languages: []string{"node"}, Tools: []protocol.ToolNeed{{Name: "rust", Choice: "tauri"}}},
	}
	d := fakeDetector{"node": {Program: "node", Version: "22.1.0"}, "rust": {Program: "cargo", Version: "1.80.0"}}

	got := SelectionStatuses(context.Background(), d, t.TempDir(), sources, "node", []string{"react", "aws-lambda", "tauri"})
	state := map[string]State{}
	for _, s := range got {
		state[s.Name] = s.State
	}
	want := map[string]State{"node": Found, "aws-sam-cli": Missing, "rust": Found}
	if len(state) != len(want) {
		t.Fatalf("tools %v, want %v (a host the stack does not use is not listed)", state, want)
	}
	for name, st := range want {
		if state[name] != st {
			t.Errorf("%s: %s, want %s", name, state[name], st)
		}
	}
}

func TestForSystemKeepsToolsOfThatSystemAndToolsWithoutAList(t *testing.T) {
	tools := []protocol.ToolNeed{
		{Name: "git"}, {Name: "xcode-clt", Systems: []string{"darwin"}}, {Name: "webview2", Systems: []string{"windows"}},
	}
	got := ForSystem(tools, "darwin")
	if len(got) != 2 || got[0].Name != "git" || got[1].Name != "xcode-clt" {
		t.Fatalf("%+v", got)
	}
	if got := ForSystem(tools, "linux"); len(got) != 1 {
		t.Fatalf("%+v", got)
	}
}
