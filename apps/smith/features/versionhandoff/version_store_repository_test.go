package versionhandoff

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspacePinIsTheNearestSmithTomlAndNewestInstalledHasItsProgram(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "smith.toml"), "[workspace]\nsmith = \"2.40.0\"\n")
	deep := filepath.Join(root, "apps", "web", "src")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := WorkspacePin(deep); got != "2.40.0" {
		t.Errorf("pin from a folder below = %q", got)
	}
	// A nearer smith.toml without a pin decides: the outer one is not read.
	write(t, filepath.Join(root, "apps", "smith.toml"), "[workspace]\nname = \"inner\"\n")
	if got := WorkspacePin(deep); got != "" {
		t.Errorf("pin under a nearer smith.toml = %q", got)
	}

	home := t.TempDir()
	if NewestInstalled(home, "linux") != "" {
		t.Error("a version is installed in an empty home")
	}
	for _, v := range []string{"2.9.0", "2.10.0", "3.0.0-beta.1"} {
		write(t, filepath.Join(home, "versions", v, "smith"), "program")
	}
	// A folder without its program, a download in progress and another name do not count.
	for _, v := range []string{"4.0.0", ".5.0.0-77", "notes"} {
		if err := os.MkdirAll(filepath.Join(home, "versions", v), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if got := NewestInstalled(home, "linux"); got != "3.0.0-beta.1" {
		t.Errorf("newest installed = %q", got)
	}
	if got := NewestInstalled(home, "windows"); got != "" {
		t.Errorf("newest installed on Windows, with no smith.exe = %q", got)
	}
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o755); err != nil {
		t.Fatal(err)
	}
}
