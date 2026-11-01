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
	dir := filepath.Join(ws, "apps/infra")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "main.bicep")
	if err := os.WriteFile(file, []byte("metadata version = '0.0.0'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "apps/infra"}, Version: "2.0.0-beta.3"})
	if err != nil || len(res.Files) != 1 {
		t.Fatal(res, err)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "metadata version = '2.0.0-beta.3'") {
		t.Errorf("want %q in %q", "metadata version = '2.0.0-beta.3'", got)
	}
	pub, err := Publish(protocol.PublishParams{Tag: "beta"})
	if err != nil || pub.Skipped == "" {
		t.Errorf("a prerelease is not published to a registry: %+v %v", pub, err)
	}
}
