package pluginpack

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func widgetDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"plugin.json":                               `{"name":"widget","version":"0.1.0","command":["./bin/smith-plugin-widget-{os}-{arch}{exe}"]}`,
		"bin/smith-plugin-widget-linux-amd64":       "elf",
		"bin/smith-plugin-widget-windows-arm64.exe": "pe",
		"bin/notes.txt":                             "not a program",
		".git/config":                               "ignored",
		"checksums.sha256":                          "stale",
		"old-0.0.1.smith-plugin.zip":                "ignored",
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

var widget = Plugin{Name: "widget", Version: "0.1.0", Command: []string{"./bin/smith-plugin-widget-{os}-{arch}{exe}"}}

func TestPackIsDeterministicAndMarksTheProgramsExecutable(t *testing.T) {
	dir := widgetDir(t)
	a, err := Pack(dir, widget, filepath.Join(t.TempDir(), "a"))
	if err != nil {
		t.Fatal(err)
	}
	// The mtime of the sources must not matter.
	old := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	for _, name := range []string{"plugin.json", "bin/notes.txt"} {
		if err := os.Chtimes(filepath.Join(dir, filepath.FromSlash(name)), old, old); err != nil {
			t.Fatal(err)
		}
	}
	b, err := Pack(dir, widget, filepath.Join(t.TempDir(), "b"))
	if err != nil {
		t.Fatal(err)
	}
	if a.SHA256 != b.SHA256 {
		t.Fatalf("two packs of the same directory differ: %s vs %s", a.SHA256, b.SHA256)
	}
	if filepath.Base(a.Path) != "widget-0.1.0.smith-plugin.zip" {
		t.Fatalf("name %s", a.Path)
	}
	want := []string{"bin/notes.txt", "bin/smith-plugin-widget-linux-amd64", "bin/smith-plugin-widget-windows-arm64.exe", "plugin.json"}
	if strings.Join(a.Files, ",") != strings.Join(want, ",") {
		t.Fatalf("files %v", a.Files)
	}
	if len(a.Programs) != 2 {
		t.Fatalf("programs %v", a.Programs)
	}
	side, _ := os.ReadFile(a.DigestPath)
	if string(side) != a.SHA256+"  widget-0.1.0.smith-plugin.zip\n" {
		t.Fatalf("sidecar %q", side)
	}

	raw, _ := os.ReadFile(a.Path)
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	modes := map[string]os.FileMode{}
	for _, f := range zr.File {
		modes[f.Name] = f.Mode().Perm()
		if f.Modified.Year() != 2000 {
			t.Errorf("%s has timestamp %v", f.Name, f.Modified)
		}
	}
	if modes["bin/smith-plugin-widget-linux-amd64"] != 0o755 || modes["bin/smith-plugin-widget-windows-arm64.exe"] != 0o755 {
		t.Errorf("programs not executable: %v", modes)
	}
	if modes["bin/notes.txt"] != 0o644 || modes["checksums.sha256"] != 0o644 {
		t.Errorf("other files: %v", modes)
	}
}

func TestPackNeedsAProgramThatMatchesTheCommand(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Pack(dir, widget, t.TempDir()); err == nil || !strings.Contains(err.Error(), "no file") {
		t.Fatalf("got %v", err)
	}
	// A program looked up on PATH is not packed, and that is fine.
	if _, err := Pack(dir, Plugin{Name: "w", Version: "1", Command: []string{"widget-tool"}}, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestPackRefusesSymbolicLinksAndBadNames(t *testing.T) {
	dir := widgetDir(t)
	if err := os.Symlink("/etc/passwd", filepath.Join(dir, "link")); err != nil {
		t.Skip("symlinks are not available here")
	}
	if _, err := Pack(dir, widget, t.TempDir()); err == nil || !strings.Contains(err.Error(), "link") {
		t.Fatalf("got %v", err)
	}
	if _, err := Pack(dir, Plugin{Name: "../x", Version: "1"}, t.TempDir()); err == nil {
		t.Fatal("packed a bad name")
	}
}
