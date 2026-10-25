package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/conformance"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func connect(t *testing.T, workspace string) *host.Client {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(&layout{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: workspace})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestConforms(t *testing.T) {
	report := conformance.Check(connect(t, t.TempDir()), conformance.Options{})
	if !report.OK() {
		t.Fatal(report.Failures())
	}
}

func TestLintWaitsForTheLayoutCheck(t *testing.T) {
	var res protocol.ContributeResult
	err := connect(t, t.TempDir()).Call(protocol.MethodContribute, protocol.ContributeParams{Projects: []protocol.Project{{Name: "core", Root: "libs/core"}}}, &res, 0)
	if err != nil {
		t.Fatal(err)
	}
	c := res.Projects["core"]
	if _, ok := c.Tasks["layout-check"]; !ok {
		t.Fatalf("no layout-check task: %+v", c)
	}
	if got := c.DependsOn["lint"]; len(got) != 1 || got[0] != "layout-check" {
		t.Errorf("lint depends on %v", got)
	}
}

func TestFindJunkDrawers(t *testing.T) {
	dir := t.TempDir()
	for _, d := range []string{"src/orders", "src/Utils", "dist/helpers", "common"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	found, err := findJunkDrawers(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 || found[0] != "common" || found[1] != "src/Utils" {
		t.Fatalf("got %v (dist is build output and is skipped)", found)
	}
}

func TestDoctorWarnsAboutAMissingReadme(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755)
	var res protocol.DoctorResult
	err := connect(t, ws).Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{{Name: "core", Root: "libs/core"}}}, &res, 0)
	if err != nil {
		t.Fatal(err)
	}
	f := res.Findings[0]
	if f.OK || !f.Warning || f.Remedy == "" {
		t.Fatalf("%+v", f)
	}
}
