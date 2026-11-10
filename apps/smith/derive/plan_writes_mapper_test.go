package derive

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPlanWritesListsOnlyWhatDeriveWrites(t *testing.T) {
	file := func(path string) protocol.DerivedFile {
		return protocol.DerivedFile{Path: path, Renderer: "node.tsconfig"}
	}
	plan := Plan{Items: []Item{
		{File: file("a/tsconfig.json"), Plugin: "node", Status: StatusMissing, Owner: "smith"},
		{File: file("b/tsconfig.json"), Plugin: "node", Status: StatusStale, Owner: "smith"},
		{File: file("c/tsconfig.json"), Plugin: "node", Status: StatusDrift, Owner: "smith"},
		{File: file("d/tsconfig.json"), Plugin: "node", Status: StatusCurrent, Owner: "smith"},
	}}
	got := PlanWrites(t.TempDir(), plan)
	if len(got.Groups) != 2 || len(got.Groups[0].Lines) != 2 || got.Groups[1].Lines[0].Path != "smith.lock" {
		t.Fatalf("%+v", got)
	}
	if got.Groups[0].Lines[0].Owner != "smith" {
		t.Fatalf("%+v", got.Groups[0].Lines[0])
	}
	if !PlanWrites(t.TempDir(), Plan{}).Empty() {
		t.Fatal("nothing to write, no plan")
	}
}
