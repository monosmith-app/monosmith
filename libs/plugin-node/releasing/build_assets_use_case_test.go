package releasing

import (
	"archive/zip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestBuildAssetsZipsTheBuildForThisPlatform(t *testing.T) {
	real := hostPlatform
	t.Cleanup(func() { hostPlatform = real })
	hostPlatform = "linux/amd64"
	ws := t.TempDir()
	for rel, body := range map[string]string{"dist/apps/driver/main.js": "run()", "dist/apps/driver/node/bin/node": "elf"} {
		path := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte(body), 0o755)
	}
	project := protocol.Project{Name: "driver", Root: "apps/driver", Kind: protocol.ProjectApp}
	spec := &protocol.AssetSpec{Name: "{product}-{os}-{arch}.{ext}", Platforms: []string{"linux/amd64"}, OSAlias: map[string]string{"linux": "linux"}, ArchAlias: map[string]string{"amd64": "x64"}, OutDir: "dist/release/driver"}

	res, err := Publish(ws, protocol.PublishParams{Project: project, Version: "1.0.0", Assets: spec})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Assets) != 1 || res.Assets[0] != "dist/release/driver/driver-linux-x64.zip" || res.AssetPlatforms[res.Assets[0]] != "linux/amd64" {
		t.Fatalf("result %+v", res)
	}
	r, err := zip.OpenReader(filepath.Join(ws, filepath.FromSlash(res.Assets[0])))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	sort.Strings(names)
	if strings.Join(names, " ") != "main.js node/bin/node" {
		t.Errorf("the zip holds the build, by forward-slash paths: %v", names)
	}

	spec.Platforms = []string{"linux/amd64", "darwin/arm64"}
	if _, err := Publish(ws, protocol.PublishParams{Project: project, Version: "1.0.0", Assets: spec}); err == nil || !strings.Contains(err.Error(), "native") || !strings.Contains(err.Error(), "Nothing was tagged") {
		t.Errorf("another platform cannot be built here, and the error says how to build it: %v", err)
	}
}

func TestBuildAssetsDryRunPlansAnotherPlatformWithANote(t *testing.T) {
	real := hostPlatform
	t.Cleanup(func() { hostPlatform = real })
	hostPlatform = "darwin/arm64"
	project := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	spec := &protocol.AssetSpec{Platforms: []string{"windows/amd64"}, OutDir: "dist/release/api"}

	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: project, Version: "6.2.0", Assets: spec, DryRun: true})
	if err != nil {
		t.Fatalf("a dry run does not fail on a platform it will not build: %v", err)
	}
	want := "dist/release/api/api_6.2.0_windows_amd64.zip"
	if len(res.Assets) != 1 || res.Assets[0] != want || res.AssetPlatforms[want] != "windows/amd64" {
		t.Errorf("the dry run names the other platform's asset: %+v", res)
	}
	if len(res.Notes) != 1 || !strings.Contains(res.Notes[0], "windows/amd64") || !strings.Contains(res.Notes[0], "native") {
		t.Errorf("the dry run says this machine cannot build it, and how to fix it: %v", res.Notes)
	}
}

func TestBuildAssetsNeedsABuild(t *testing.T) {
	real := hostPlatform
	t.Cleanup(func() { hostPlatform = real })
	hostPlatform = "windows/amd64"
	spec := &protocol.AssetSpec{Platforms: []string{"windows/amd64"}, OutDir: "dist/release/driver"}
	project := protocol.Project{Name: "driver", Root: "apps/driver"}
	dry, err := Publish(t.TempDir(), protocol.PublishParams{Project: project, Version: "1.0.0", Assets: spec, DryRun: true})
	if err != nil || dry.Assets[0] != "dist/release/driver/driver_1.0.0_windows_amd64.zip" {
		t.Errorf("a dry run names the asset with the default pattern: %+v %v", dry, err)
	}
	if _, err := Publish(t.TempDir(), protocol.PublishParams{Project: project, Version: "1.0.0", Assets: spec}); err == nil || !strings.Contains(err.Error(), "build it first") {
		t.Errorf("no build: %v", err)
	}
}
