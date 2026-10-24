package corefixes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUntrackFixRemovesIgnoredFilesWithoutAQuestion(t *testing.T) {
	fix := UntrackFix(t.TempDir(), []TrackedFile{
		{Path: "b/x.tsbuildinfo", Ignored: true},
		{Path: "a/y.tsbuildinfo", Ignored: true},
	})
	if fix == nil || fix.Safe || len(fix.Choices) != 0 || len(fix.Steps) != 1 {
		t.Fatalf("one step, nothing to ask, never safe: %+v", fix)
	}
	got := strings.Join(fix.Steps[0].Argv, " ")
	if got != "git rm -r --cached --ignore-unmatch --quiet -- a/y.tsbuildinfo b/x.tsbuildinfo" {
		t.Fatalf("%s", got)
	}
}

func TestUntrackFixAsksWhetherToIgnoreOutputDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("node_modules"), 0o644); err != nil {
		t.Fatal(err)
	}
	fix := UntrackFix(root, []TrackedFile{
		{Path: "dist/apps/web/a.js", OutputDir: "dist/apps/web"},
		{Path: "dist/apps/web/b.js", OutputDir: "dist/apps/web"},
		{Path: "dist/apps/web/c.js", Ignored: true},
	})
	if len(fix.Steps) != 1 || !strings.HasSuffix(strings.Join(fix.Steps[0].Argv, " "), "-- dist/apps/web") {
		t.Fatalf("the directory is removed whole, its files need no step: %+v", fix.Steps)
	}
	if len(fix.Choices) != 1 || len(fix.Choices[0].Options) != 2 {
		t.Fatalf("one question, two answers: %+v", fix.Choices)
	}
	with := fix.Choices[0].Options[0].Fix.Files
	if len(with) != 1 || with[0].Path != ".gitignore" || !strings.HasPrefix(with[0].Content, "node_modules\n") || !strings.Contains(with[0].Content, "/dist/apps/web/\n") {
		t.Fatalf("the rule is appended to the existing .gitignore: %+v", with)
	}
	if only := fix.Choices[0].Options[1].Fix; len(only.Files) != 0 || len(only.Steps) != 0 {
		t.Fatalf("the second answer changes nothing more: %+v", only)
	}
	if UntrackFix(root, nil) != nil {
		t.Fatal("no files, no fix")
	}
}

func TestUntrackFixSplitsALongListAcrossSteps(t *testing.T) {
	var files []TrackedFile
	for i := 0; i < 250; i++ {
		files = append(files, TrackedFile{Path: filepath.ToSlash(filepath.Join("gen", string(rune('a'+i%26)), strings.Repeat("x", i%7+1)+string(rune('0'+i/26)))), Ignored: true})
	}
	fix := UntrackFix(t.TempDir(), files)
	if len(fix.Steps) < 2 {
		t.Fatalf("a command line has a limit: %d steps", len(fix.Steps))
	}
}
