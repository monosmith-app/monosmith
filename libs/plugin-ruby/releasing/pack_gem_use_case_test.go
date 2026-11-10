package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPackBuildsTheGemAndPushesNothing(t *testing.T) {
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, "packages", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "packages", "core", "acme_core.gemspec"), []byte("spec.version = '1.0.0'"), 0o644); err != nil {
		t.Fatal(err)
	}
	var ran []string
	real := GemRunner
	defer func() { GemRunner = real }()
	GemRunner = func(_ string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		return os.WriteFile(argv[len(argv)-1], []byte("gem"), 0o644)
	}
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "core", Kind: protocol.ProjectPackage, Root: "packages/core", Exports: []string{"acme_core"}}, Version: "1.0.0", PackDir: "dist/release/core/packages"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || len(res.Assets) != 1 || res.Assets[0] != "dist/release/core/packages/acme_core-1.0.0.gem" {
		t.Fatalf("result %+v", res)
	}
	if len(ran) != 1 || !strings.HasPrefix(ran[0], "gem build ") {
		t.Fatalf("only gem build runs: %v", ran)
	}
}

func TestPackDryRunNamesTheGem(t *testing.T) {
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Name: "core", Kind: protocol.ProjectPackage}, Version: "2.0.0-beta.1", DryRun: true, PackDir: "out"})
	if err != nil || res.Skipped != "dry run" || len(res.Assets) != 1 || res.Assets[0] != "out/core-2.0.0.pre.beta.1.gem" {
		t.Fatalf("res %+v err %v", res, err)
	}
}
