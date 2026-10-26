package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckVenvIgnored(t *testing.T) {
	real := checkIgnore
	defer func() { checkIgnore = real }()

	ws := t.TempDir()
	if _, ok := CheckVenvIgnored(ws); ok {
		t.Fatal("no venv, nothing to say")
	}
	_ = os.MkdirAll(filepath.Join(ws, ".venv"), 0o755)

	checkIgnore = func(string) int { return 128 }
	if _, ok := CheckVenvIgnored(ws); ok {
		t.Fatal("outside a git repository there is nothing to say")
	}

	checkIgnore = func(string) int { return 0 }
	if f, ok := CheckVenvIgnored(ws); !ok || !f.OK {
		t.Fatalf("an ignored venv passes: %+v", f)
	}

	checkIgnore = func(string) int { return 1 }
	_ = os.WriteFile(filepath.Join(ws, ".gitignore"), []byte("dist/\n"), 0o644)
	f, ok := CheckVenvIgnored(ws)
	if !ok || f.OK || f.Remedy == "" || f.Fix == nil || !f.Fix.Safe {
		t.Fatalf("an un-ignored venv fails with a safe fix: %+v", f)
	}
	if content := f.Fix.Files[0].Content; f.Fix.Files[0].Path != ".gitignore" || !strings.Contains(content, "dist/") || !strings.Contains(content, ".venv/") {
		t.Fatalf("fix: %q", content)
	}
}

func TestCheckVenvIgnoredAskingGit(t *testing.T) {
	ws := t.TempDir()
	init := exec.Command("git", "init", "-q")
	init.Dir = ws
	if err := init.Run(); err != nil {
		t.Skip("git is not available")
	}
	_ = os.MkdirAll(filepath.Join(ws, ".venv"), 0o755)
	if f, ok := CheckVenvIgnored(ws); !ok || f.OK {
		t.Fatalf("git sees an un-ignored venv: %+v", f)
	}
	_ = os.WriteFile(filepath.Join(ws, ".gitignore"), []byte(".venv/\n"), 0o644)
	if f, ok := CheckVenvIgnored(ws); !ok || !f.OK {
		t.Fatalf("git sees an ignored venv: %+v", f)
	}
}
