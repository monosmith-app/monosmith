package releasing

import (
	"archive/zip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestBuildAssetsWithRealCargo builds a tiny app for this machine's platform
// with the real cargo, unpacks the archive and runs the program in it. It is
// skipped where cargo, or the target of this machine, is not installed.
func TestBuildAssetsWithRealCargo(t *testing.T) {
	platform := runtime.GOOS + "/" + runtime.GOARCH
	triple, err := Target(platform)
	if _, lookErr := exec.LookPath("cargo"); lookErr != nil || err != nil {
		t.Skip("cargo is not installed, or Rust has no target for this machine")
	}
	if installed, ok := InstalledTargets(t.TempDir()); ok && !slices.Contains(installed, triple) {
		t.Skip(triple + " is not installed")
	}
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "hello")
	_ = os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "Cargo.toml"), []byte("[workspace]\nresolver = \"2\"\nmembers = [\"apps/hello\"]\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"acme-hello\"\nversion = \"3.4.5\"\nedition = \"2021\"\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "src", "main.rs"), []byte("fn main() { println!(\"hello {}\", env!(\"CARGO_PKG_VERSION\")); }\n"), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "LICENSE"), []byte("MIT"), 0o644)

	project := protocol.Project{Name: "hello", Root: "apps/hello", Kind: protocol.ProjectApp, Exports: []string{"acme-hello"}}
	res, err := BuildAssets(ws, protocol.PublishParams{Project: project, Version: "3.4.5", Assets: spec(platform)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Binary != "acme-hello" || len(res.Assets) != 2 {
		t.Fatalf("%+v", res)
	}
	zr, err := zip.OpenReader(filepath.Join(ws, filepath.FromSlash(res.Assets[0])))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	out := t.TempDir()
	var program string
	for _, f := range zr.File {
		in, _ := f.Open()
		path := filepath.Join(out, f.Name)
		dst, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o755)
		_, _ = io.Copy(dst, in)
		_ = dst.Close()
		_ = in.Close()
		if strings.HasPrefix(f.Name, "acme-hello") {
			program = path
		}
	}
	if program == "" {
		t.Fatalf("no program in the archive of %d files", len(zr.File))
	}
	got, err := exec.Command(program).Output()
	if err != nil || strings.TrimSpace(string(got)) != "hello 3.4.5" {
		t.Fatalf("%q %v", got, err)
	}
}
