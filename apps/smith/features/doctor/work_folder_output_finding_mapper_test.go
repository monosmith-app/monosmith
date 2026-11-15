package doctor

import (
	"strings"
	"testing"
)

func TestSharedWorkOutputFindingPassesWithNothingToSay(t *testing.T) {
	if f := sharedWorkOutputFinding(nil); !f.OK || f.Detail != "" || f.Check != sharedWorkOutputCheck {
		t.Fatalf("%+v", f)
	}
}

func TestSharedWorkOutputFindingFailsAndSaysWhy(t *testing.T) {
	f := sharedWorkOutputFinding([]string{"cli:build caches dist/cargo, where api ships from"})
	if f.OK || f.Warning || f.Problem != "cli:build caches dist/cargo, where api ships from" ||
		!strings.Contains(f.Detail, "empties each output folder first") || !strings.Contains(f.Remedy, "dist/<project dir>") {
		t.Fatalf("%+v", f)
	}
}
