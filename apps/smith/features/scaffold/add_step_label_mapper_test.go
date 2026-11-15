package scaffold

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAddStepsNameWhatIsHappeningInWords(t *testing.T) {
	st := protocol.Stack{Name: "node-jest", Description: "TypeScript on Node, linted by eslint, tested with Jest"}
	got := createStep("package", "demo-pkg", "packages/demo-pkg", st, "node")
	if got != "Creating package demo-pkg in packages/demo-pkg: TypeScript on Node, linted by eslint, tested with Jest" {
		t.Fatalf("create: %q", got)
	}
	if strings.Contains(got, "stack node-jest") {
		t.Fatal("the stack's name means nothing to most people")
	}
	if got := createStep("lib", "core", "libs/core", protocol.Stack{Name: protocol.NoStack}, "go"); got != "Creating lib core in libs/core: go" {
		t.Fatalf("no stack falls back to the language: %q", got)
	}
	wrote := protocol.GenerateResult{Files: []string{"libs/core/eslint.config.mjs"}}
	for plugin, want := range map[string]string{"eslint": "Added linting (eslint)", "jest": "Added tests (jest)", "widget-lint": "Added widget-lint"} {
		if got := partStep(plugin, "lib", wrote); got != want {
			t.Errorf("%s: %q", plugin, got)
		}
	}
	if got := partStep("jest", "app", protocol.GenerateResult{InstallPending: true}); got != "Added tests (jest)" {
		t.Errorf("a part that only leaves an install did add something: %q", got)
	}
	if got := partStep("docker", "lib", protocol.GenerateResult{}); got != "docker: nothing to add for a lib" {
		t.Errorf("a part that changed nothing must say so: %q", got)
	}
	if got := doneStep("package", "demo-pkg", 11, "9.4s"); got != "Done: package demo-pkg is ready (11 files written, 9.4s)." {
		t.Fatalf("done: %q", got)
	}
}

func TestDirectoryReasonSaysWhyTheProjectLeftItsKindFolder(t *testing.T) {
	got := directoryReason("lib", "libs", "packages")
	if !strings.Contains(got, "goes to packages/, not libs/") || !strings.Contains(got, "list packages/* and not libs/*") {
		t.Errorf("got %q", got)
	}
	if directoryReason("lib", "libs", "libs") != "" {
		t.Error("a project in its own folder needs no reason")
	}
}
