package linuxpackages

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, dir, name, file string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary of " + name))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
}

func TestBuildPackagesEveryLinuxArchive(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dist", "release", "tool")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	archive(t, dir, "tool_1.0.0_linux_amd64.zip", "tool")
	archive(t, dir, "tool_1.0.0_linux_arm64.zip", "tool")
	archive(t, dir, "tool_1.0.0_windows_amd64.zip", "tool.exe")
	if err := os.WriteFile(filepath.Join(dir, "checksums.txt"), []byte("abc  tool_1.0.0_linux_amd64.zip\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	built := Built{
		Assets: []string{"dist/release/tool/tool_1.0.0_linux_amd64.zip", "dist/release/tool/tool_1.0.0_linux_arm64.zip", "dist/release/tool/tool_1.0.0_windows_amd64.zip", "dist/release/tool/checksums.txt"},
		Platforms: map[string]string{
			"dist/release/tool/tool_1.0.0_linux_amd64.zip":   "linux/amd64",
			"dist/release/tool/tool_1.0.0_linux_arm64.zip":   "linux/arm64",
			"dist/release/tool/tool_1.0.0_windows_amd64.zip": "windows/amd64",
		},
	}
	r := Release{Kinds: []string{"deb", "rpm"}, Product: "tool", Binary: "tool", Version: "1.0.0"}
	written, err := Build(root, built, r)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist/release/tool/tool_1.0.0_amd64.deb", "dist/release/tool/tool-1.0.0-1.x86_64.rpm", "dist/release/tool/tool_1.0.0_arm64.deb", "dist/release/tool/tool-1.0.0-1.aarch64.rpm"}
	if strings.Join(written, ",") != strings.Join(want, ",") {
		t.Fatalf("written = %v", written)
	}
	var names []string
	for _, w := range written {
		names = append(names, filepath.Base(w))
	}
	expected := Expected(r, []string{"linux/amd64", "linux/arm64", "windows/amd64", "darwin/arm64"})
	if len(expected) != 4 {
		t.Errorf("Expected = %v", expected)
	}
	for _, n := range names {
		found := false
		for _, e := range expected {
			found = found || e == n
		}
		if !found {
			t.Errorf("Expected lacks %s: %v", n, expected)
		}
	}
	sums, _ := os.ReadFile(filepath.Join(dir, "checksums.txt"))
	deb, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(want[0])))
	line := fmt.Sprintf("%x  tool_1.0.0_amd64.deb\n", sha256.Sum256(deb))
	if !strings.HasPrefix(string(sums), "abc  tool_1.0.0_linux_amd64.zip\n") || !strings.Contains(string(sums), line) || strings.Count(string(sums), "\n") != 5 {
		t.Errorf("checksums.txt = %q", sums)
	}
	// A second build (a resumed release) lists each package once.
	if _, err := Build(root, built, r); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.ReadFile(filepath.Join(dir, "checksums.txt")); strings.Count(string(again), "\n") != 5 {
		t.Errorf("checksums.txt after a rebuild = %q", again)
	}
}

func TestBuildNeedsALinuxArchive(t *testing.T) {
	if _, err := Build(t.TempDir(), Built{Platforms: map[string]string{"x.zip": "darwin/arm64"}}, Release{Kinds: []string{"deb"}, Product: "x", Binary: "x", Version: "1.0.0"}); err == nil {
		t.Error("no error without a linux archive")
	}
}
