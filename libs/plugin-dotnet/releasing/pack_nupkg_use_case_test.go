package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPackMakesTheNupkgAndPushesNothing(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "packages", "core", "Acme.Core.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	var ran []string
	real := DotnetRunner
	defer func() { DotnetRunner = real }()
	DotnetRunner = func(_ string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		return os.WriteFile(filepath.Join(ws, "dist", "release", "core", "packages", "Acme.Core.1.0.0.nupkg"), []byte("nupkg"), 0o644)
	}
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "core", Kind: protocol.ProjectPackage, Root: "packages/core", Exports: []string{"Acme.Core"}}, Version: "1.0.0", PackDir: "dist/release/core/packages"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || len(res.Assets) != 1 || res.Assets[0] != "dist/release/core/packages/Acme.Core.1.0.0.nupkg" {
		t.Fatalf("result %+v", res)
	}
	if len(ran) != 1 || !strings.HasPrefix(ran[0], "dotnet pack ") || !strings.Contains(ran[0], "-p:Version=1.0.0") {
		t.Fatalf("only dotnet pack runs: %v", ran)
	}
}

func TestPackDryRunNamesTheNupkg(t *testing.T) {
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Name: "core", Kind: protocol.ProjectPackage, Exports: []string{"Acme.Core"}}, Version: "1.0.0", DryRun: true, PackDir: "out"})
	if err != nil || res.Skipped != "dry run" || len(res.Assets) != 1 || res.Assets[0] != "out/Acme.Core.1.0.0.nupkg" {
		t.Fatalf("res %+v err %v", res, err)
	}
}
