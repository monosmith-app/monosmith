package derive

import (
	"io"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func checkItem(status Status) Item {
	return Item{File: protocol.DerivedFile{Path: "web/tsconfig.lib.json"}, Status: status}
}

func TestCheckPlanNamesTheChoiceForAHandEdit(t *testing.T) {
	err := CheckPlan(io.Discard, Plan{Items: []Item{checkItem(StatusDrift)}})
	if err == nil || !strings.Contains(err.Error(), "--eject") || strings.Contains(err.Error(), "run smith derive,") {
		t.Fatalf("a drifted file needs a choice, not smith derive: %v", err)
	}
}

func TestCheckPlanTellsAStaleFileToRunDerive(t *testing.T) {
	err := CheckPlan(io.Discard, Plan{Items: []Item{checkItem(StatusStale)}})
	if err == nil || !strings.Contains(err.Error(), "run smith derive") {
		t.Fatalf("a stale file is written by smith derive: %v", err)
	}
	if CheckPlan(io.Discard, Plan{Items: []Item{checkItem(StatusCurrent)}}) != nil {
		t.Error("a current file passes")
	}
}
