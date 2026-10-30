package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "report")
	_ = os.MkdirAll(dir, 0o755)
	main := filepath.Join(dir, "main.bas")
	const before = "Attribute VB_Name = \"Main\"\r\n' smithVersion: 0.0.0\r\nOption Explicit\r\n"
	_ = os.WriteFile(main, []byte(before), 0o644)
	project := protocol.Project{Root: "apps/report"}

	res, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0", DryRun: true})
	if err != nil || res.Files[0] != "apps/report/main.bas" {
		t.Fatal(res, err)
	}
	if got, _ := os.ReadFile(main); string(got) != before {
		t.Errorf("a dry run writes nothing: %q", got)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(main); string(got) != "Attribute VB_Name = \"Main\"\r\n' smithVersion: 0.1.0\r\nOption Explicit\r\n" {
		t.Errorf("the CRLF endings are kept: %q", got)
	}
	if res, _ := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); len(res.Files) != 0 {
		t.Errorf("the same version is no change: %+v", res)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "apps/none"}, Version: "1.0.0"}); err == nil {
		t.Error("a missing directory is an error")
	}
	_ = os.WriteFile(main, []byte("not a module"), 0o644)
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "1.0.0"}); err == nil {
		t.Error("a directory without a VBA module is an error")
	}
	if res, err := Publish(protocol.PublishParams{}); err != nil || res.Skipped == "" {
		t.Errorf("%+v %v", res, err)
	}
}
