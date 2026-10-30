package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "libs", "core")
	_ = os.MkdirAll(dir, 0o755)
	file := filepath.Join(dir, "Package.swift")
	const before = "// swift-tools-version: 5.9\n// smithVersion: 0.0.0\nimport PackageDescription\n"
	_ = os.WriteFile(file, []byte(before), 0o644)
	project := protocol.Project{Root: "libs/core"}

	res, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0", DryRun: true})
	if err != nil || res.Files[0] != "libs/core/Package.swift" {
		t.Fatal(res, err)
	}
	if got, _ := os.ReadFile(file); string(got) != before {
		t.Errorf("a dry run writes nothing: %q", got)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(file); string(got) != "// swift-tools-version: 5.9\n// smithVersion: 0.1.0\nimport PackageDescription\n" {
		t.Errorf("%q", got)
	}
	if res, _ := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); len(res.Files) != 0 {
		t.Errorf("the same version is no change: %+v", res)
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/none"}, Version: "1.0.0"}); err == nil {
		t.Error("a project without a manifest is an error")
	}
	if res, err := Publish(protocol.PublishParams{}); err != nil || res.Skipped == "" {
		t.Errorf("%+v %v", res, err)
	}
}
