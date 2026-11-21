package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestMissingTests(t *testing.T) {
	got := MissingTests([]protocol.Project{
		{Name: "tested", Tasks: map[string]protocol.Task{"test": {}}},
		{Name: "file-analyzer", Tasks: map[string]protocol.Task{"lint": {}}},
	})
	if len(got) != 1 {
		t.Fatalf("one warning for the project without a test task: %+v", got)
	}
	f := got[0]
	if f.Project != "file-analyzer" || !f.Warning || f.OK {
		t.Fatalf("a warning about file-analyzer: %+v", f)
	}
	if !strings.Contains(f.Problem, "has no tests") || !strings.Contains(f.Remedy, "smith run test -p file-analyzer") {
		t.Fatalf("the problem and the remedy: %+v", f)
	}
}
