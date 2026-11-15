package depsrepair

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDeletePathAcceptsAPathOfATool(t *testing.T) {
	root := t.TempDir()
	got, err := checkDeletePath(root, "apps/web/node_modules", []string{"apps/web"})
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "apps", "web", "node_modules"); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestCheckDeletePathRefuses(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{
		"", ".", "..", "../x", "a/../../x", "/etc", `C:\x`, `a\b`, "a//b", "./a",
		".git", ".git/hooks", ".smith", "smith.toml",
		"apps/web", "apps", // a project's root and a folder that holds one
	} {
		if _, err := checkDeletePath(root, rel, []string{".", "apps/web"}); err == nil {
			t.Errorf("%q is accepted", rel)
		}
	}
}

// A path below a link would delete through the link, out of the workspace.
func TestCheckDeletePathRefusesAPathBelowALink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outside, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	link(t, outside, filepath.Join(root, "vendor"))
	_, err := checkDeletePath(root, "vendor/node_modules", nil)
	if err == nil || !strings.Contains(err.Error(), "link") {
		t.Fatalf("want a refusal that names the link, got %v", err)
	}
	// The link itself may be deleted: as a link.
	if _, err := checkDeletePath(root, "vendor", nil); err != nil {
		t.Fatalf("the link itself is a path smith deletes as a link: %v", err)
	}
}
