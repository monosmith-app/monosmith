package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A release channel stamps a prerelease version: the manifest keeps it in the
// spelling the ecosystem reads, and Publish still moves no registry pointer.
func TestWriteVersionStampsAChannelPrerelease(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "libs/core")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "Package.swift")
	if err := os.WriteFile(file, []byte("// swift-tools-version: 5.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/core"}, Version: "2.0.0-beta.3"})
	if err != nil || len(res.Files) != 1 {
		t.Fatal(res, err)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "// smithVersion: 2.0.0-beta.3") {
		t.Errorf("want %q in %q", "// smithVersion: 2.0.0-beta.3", got)
	}
	pub, err := Publish(protocol.PublishParams{Tag: "beta"})
	if err != nil || pub.Skipped == "" {
		t.Errorf("a prerelease is not published to a registry: %+v %v", pub, err)
	}
}
