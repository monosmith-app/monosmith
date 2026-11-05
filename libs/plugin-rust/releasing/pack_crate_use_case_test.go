package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPackWritesTheCrateAndPublishesNothing(t *testing.T) {
	ws := t.TempDir()
	var ran []string
	real := CargoRunner
	defer func() { CargoRunner = real }()
	CargoRunner = func(_ string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		target := argv[len(argv)-1]
		if err := os.MkdirAll(filepath.Join(target, "package"), 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(target, "package", "acme-1.0.0.crate"), []byte("crate"), 0o644)
	}
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "acme", Kind: protocol.ProjectPackage, Root: "crates/acme", Exports: []string{"acme"}}, Version: "1.0.0", PackDir: "dist/release/acme/packages"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || len(res.Assets) != 1 || res.Assets[0] != "dist/release/acme/packages/acme-1.0.0.crate" {
		t.Fatalf("result %+v", res)
	}
	if len(ran) != 1 || !strings.HasPrefix(ran[0], "cargo package -p acme") || strings.Contains(ran[0], "publish") {
		t.Fatalf("only cargo package runs: %v", ran)
	}
	if raw, err := os.ReadFile(filepath.Join(ws, "dist", "release", "acme", "packages", "acme-1.0.0.crate")); err != nil || string(raw) != "crate" {
		t.Fatalf("the crate is in the pack folder: %v", err)
	}
}

func TestPackDryRunNamesTheCrateAndRunsNothing(t *testing.T) {
	real := CargoRunner
	defer func() { CargoRunner = real }()
	CargoRunner = func(string, []string) error { t.Fatal("a dry run runs no cargo"); return nil }
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Name: "acme", Kind: protocol.ProjectPackage}, Version: "1.0.0", DryRun: true, PackDir: "out"})
	if err != nil || res.Skipped != "dry run" || len(res.Assets) != 1 || res.Assets[0] != "out/acme-1.0.0.crate" {
		t.Fatalf("res %+v err %v", res, err)
	}
}
