package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "notes")
	_ = os.MkdirAll(dir, 0o755)
	main := filepath.Join(dir, "main.applescript")
	_ = os.WriteFile(main, []byte("-- smithVersion: 0.0.0\non run\nend run\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "a.applescript"), []byte("on x()\nend x\n"), 0o644)
	project := protocol.Project{Root: "apps/notes"}

	res, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0", DryRun: true})
	if err != nil || res.Files[0] != "apps/notes/main.applescript" {
		t.Fatal(res, err)
	}
	if got, _ := os.ReadFile(main); string(got) != "-- smithVersion: 0.0.0\non run\nend run\n" {
		t.Errorf("a dry run writes nothing: %q", got)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(main); string(got) != "-- smithVersion: 0.1.0\non run\nend run\n" {
		t.Errorf("%q", got)
	}
	if res, _ := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); len(res.Files) != 0 {
		t.Errorf("the same version is no change: %+v", res)
	}

	_ = os.WriteFile(main, []byte("on run\nend run\n"), 0o644)
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(main); string(got) != "-- smithVersion: 1.0.0\non run\nend run\n" {
		t.Errorf("a missing comment is added on top: %q", got)
	}

	if _, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "apps/none"}, Version: "1.0.0"}); err == nil {
		t.Error("a missing directory is an error")
	}
	if res, err := Publish(protocol.PublishParams{}); err != nil || res.Skipped == "" {
		t.Errorf("%+v %v", res, err)
	}
}
