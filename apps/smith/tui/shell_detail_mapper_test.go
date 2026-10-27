package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
)

func TestATaskHintSaysWhatRunsFirstWithoutNxSyntax(t *testing.T) {
	hint := taskHint(projectlist.Task{Name: "package", DependsOn: []string{"^build", "build"}, Cache: true})
	if strings.Contains(hint, "^") || !strings.Contains(hint, "runs after the build of the projects it needs and its own build") {
		t.Fatalf("%s", hint)
	}
}

func TestTheStackRowSaysWhatTheStackGives(t *testing.T) {
	if got := stackText("node-jest"); got != "node-jest (ESLint and Jest tasks)" {
		t.Fatalf("%s", got)
	}
	if got := stackText("custom"); got != "custom" {
		t.Fatalf("%s", got)
	}
	if taskHint(projectlist.Task{Name: "deps-check"}) == "a task of this project" {
		t.Fatal("deps-check is unexplained")
	}
}
