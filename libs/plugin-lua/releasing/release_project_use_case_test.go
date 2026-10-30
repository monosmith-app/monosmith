package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	out, ok := SetVersion("package = \"x\"\nversion = \"0.0.0-4\"\n", "1.1.0")
	if !ok || out != "package = \"x\"\nversion = \"1.1.0-1\"\n" {
		t.Errorf("%q %v", out, ok)
	}
	if _, ok := SetVersion(out, "1.1.0"); ok {
		t.Error("the same version is no change")
	}
	out, ok = SetVersion("package = 'x'\nsource = {}\n", "2.0.0")
	if !ok || out != "package = 'x'\nversion = \"2.0.0-1\"\nsource = {}\n" {
		t.Errorf("a missing version is added under package: %q %v", out, ok)
	}
	if _, ok := SetVersion("-- nothing\n", "1.0.0"); ok {
		t.Error("without a package there is nowhere to add it")
	}
	out, _ = SetVersion("version = '0.1.0-1'\n", "0.2.0")
	if out != "version = '0.2.0-1'\n" {
		t.Errorf("quote style is kept: %q", out)
	}
}

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "libs", "x")
	_ = os.MkdirAll(dir, 0o755)
	file := filepath.Join(dir, "x-0.0.0-1.rockspec")
	_ = os.WriteFile(file, []byte("package = \"x\"\nversion = \"0.0.0-1\"\n"), 0o644)
	project := protocol.Project{Root: "libs/x"}
	res, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0", DryRun: true})
	if err != nil || res.Files[0] != "libs/x/x-0.0.0-1.rockspec" {
		t.Fatal(res, err)
	}
	if got, _ := os.ReadFile(file); string(got) != "package = \"x\"\nversion = \"0.0.0-1\"\n" {
		t.Errorf("a dry run writes nothing: %q", got)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(file); string(got) != "package = \"x\"\nversion = \"0.1.0-1\"\n" {
		t.Errorf("%q", got)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/none"}, Version: "1.0.0"}); err == nil {
		t.Error("a project without a rockspec is an error")
	}
	if res, err := Publish(protocol.PublishParams{}); err != nil || res.Skipped == "" {
		t.Errorf("%+v %v", res, err)
	}
}
