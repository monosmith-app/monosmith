package derive

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func wholeFile(renderer, content string) protocol.DerivedFile {
	return protocol.DerivedFile{Path: "a/tsconfig.json", Mode: protocol.DeriveFile, Renderer: renderer, Version: 1, Content: content}
}

func TestTwoRenderersThatAgreeOnAWholeFileAreNotAConflict(t *testing.T) {
	plan, err := PlanFor(t.TempDir(), derivelock.Record{}, []Proposal{
		{Plugin: "node", File: wholeFile("node.sibling-shape", "{}\n")},
		{Plugin: "typescript", File: wholeFile("typescript.tsconfig", "{}\n")},
	})
	if err != nil || len(plan.Items) != 1 {
		t.Fatalf("%v %+v", err, plan.Items)
	}
}

func TestTheRendererTheRecordListsWinsADisputeOverAWholeFile(t *testing.T) {
	rec := derivelock.Record{Files: []derivelock.Entry{{Path: "a/tsconfig.json", Owner: derivelock.OwnerSmith, Renderer: "node.sibling-shape", Mode: protocol.DeriveFile}}}
	plan, err := PlanFor(t.TempDir(), rec, []Proposal{
		{Plugin: "typescript", File: wholeFile("typescript.tsconfig", "{\"x\":1}\n")},
		{Plugin: "node", File: wholeFile("node.sibling-shape", "{}\n")},
	})
	if err != nil || len(plan.Items) != 1 || plan.Items[0].Plugin != "node" {
		t.Fatalf("%v %+v", err, plan.Items)
	}
}

func TestTwoRenderersThatDisagreeOnAWholeFileAreAConflict(t *testing.T) {
	_, err := PlanFor(t.TempDir(), derivelock.Record{}, []Proposal{
		{Plugin: "node", File: wholeFile("a", "{}\n")},
		{Plugin: "typescript", File: wholeFile("b", "{\"x\":1}\n")},
	})
	if err == nil {
		t.Fatal("two plugins own the same whole file with different content")
	}
}
