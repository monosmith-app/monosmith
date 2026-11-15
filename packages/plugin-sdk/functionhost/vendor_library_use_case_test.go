package functionhost

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// npm links a workspace on Windows as a junction, which Go does not report as a
// symbolic link: a test that links with os.Symlink would not catch it.
func TestLinkedFolderFollowsAWindowsJunction(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("junctions exist on Windows only")
	}
	ws := t.TempDir()
	lib := filepath.Join(ws, "libs", "models")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "package.json"), []byte(`{"name":"@quotes/models"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	scope := filepath.Join(ws, "node_modules", "@quotes")
	if err := os.MkdirAll(scope, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", filepath.Join(scope, "models"), lib).CombinedOutput(); err != nil {
		t.Fatalf("mklink: %v %s", err, out)
	}
	app := filepath.Join(ws, "apps", "api")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := linkedFolder(app, "@quotes/models"); got != lib {
		t.Fatalf("the junction leads to the library: got %q, want %q", got, lib)
	}
	if got := linkedFolder(app, "@quotes/other"); got != "" {
		t.Fatalf("no link, no library: %q", got)
	}
}
