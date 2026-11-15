package taskrun

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFailureHintNamesTheMissingWorkspaceFile(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "declaration-specifiers.mjs")
	output := "node:internal/modules/cjs/loader:1424\n  throw err;\nError: Cannot find module '" + missing + "'\n    at Module._resolveFilename"
	hint := FailureHint(output, root)
	if !strings.Contains(hint, "needs declaration-specifiers.mjs") || !strings.Contains(hint, "git checkout -- declaration-specifiers.mjs") {
		t.Fatalf("%q", hint)
	}
}

func TestFailureHintOtherCauses(t *testing.T) {
	root := t.TempDir()
	for output, want := range map[string]string{
		"Error: Cannot find module 'left-pad'": "package left-pad is not installed",
		"sh: 1: tsc: not found":                "command tsc was not found",
		"sh: smith: command not found":         "monosmith",
		"'esbuild' is not recognized as an internal or external command, operable program or batch file.": "command esbuild was not found",
		"1 test failed": "",
	} {
		if hint := FailureHint(output, root); (want == "" && hint != "") || !strings.Contains(hint, want) {
			t.Errorf("%q: got %q, want it to contain %q", output, hint, want)
		}
	}
}
