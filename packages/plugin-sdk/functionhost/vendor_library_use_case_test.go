package functionhost

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func TestPlaceLibraryCodeShipsATypeScriptLibraryBuiltInItsOwnFolder(t *testing.T) {
	root := t.TempDir()
	folder := filepath.Join(root, "libs", "core")
	write(t, filepath.Join(folder, "dist", "index.js"), "built")
	target := filepath.Join(t.TempDir(), "core")
	manifest := map[string]any{"name": "@acme/core", "main": "src/index.ts", "exports": "./src/index.ts"}
	builds := []string{filepath.Join(folder, "dist"), filepath.Join(root, "dist", "libs", "core")}
	if err := placeLibraryCode("@acme/core", "src/index.ts", folder, builds, target, manifest); err != nil {
		t.Fatalf("a library built in its own dist (ADR 0102) ships: %v", err)
	}
	if manifest["main"] != "./dist/index.js" || manifest["exports"] != nil {
		t.Errorf("main points at the built file and exports goes: %v", manifest)
	}
	if _, err := os.Stat(filepath.Join(target, "dist", "index.js")); err != nil {
		t.Errorf("the build is copied: %v", err)
	}
}

func TestPlaceLibraryCodeStillShipsTheOldWorkspaceDistLayout(t *testing.T) {
	root := t.TempDir()
	folder := filepath.Join(root, "libs", "core")
	write(t, filepath.Join(root, "dist", "libs", "core", "index.js"), "built")
	target := filepath.Join(t.TempDir(), "core")
	manifest := map[string]any{"name": "@acme/core"}
	builds := []string{filepath.Join(folder, "dist"), filepath.Join(root, "dist", "libs", "core")}
	if err := placeLibraryCode("@acme/core", "./src/index.ts", folder, builds, target, manifest); err != nil {
		t.Fatalf("a build in the workspace's dist/<library dir> ships: %v", err)
	}
	if err := placeLibraryCode("@acme/none", "src/index.ts", filepath.Join(root, "libs", "none"), []string{filepath.Join(root, "libs", "none", "dist")}, target, map[string]any{}); err == nil || !strings.Contains(err.Error(), "has no build in") {
		t.Errorf("no build anywhere: %v", err)
	}
}
