package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
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
	go func() { _ = serve.Serve(&bundle{}, pluginR, pluginW); _ = pluginW.Close() }()
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

func TestOnlyAppsGetABundleTask(t *testing.T) {
	var res protocol.ContributeResult
	projects := []protocol.Project{{Name: "shop", Root: "apps/shop", Kind: protocol.ProjectApp}, {Name: "core", Root: "libs/core", Kind: protocol.ProjectLib}}
	if err := connect(t, t.TempDir()).Call(protocol.MethodContribute, protocol.ContributeParams{Projects: projects}, &res, 0); err != nil {
		t.Fatal(err)
	}
	if _, ok := res.Projects["shop"].Tasks["bundle"]; !ok {
		t.Error("the app has no bundle task")
	}
	if _, ok := res.Projects["core"]; ok {
		t.Error("a lib must not be contributed to: smith would treat the plugin as its deployer")
	}
}

func TestPublishIsIdempotentAndHonoursDryRun(t *testing.T) {
	ws := t.TempDir()
	built := filepath.Join(ws, "dist", "apps", "shop")
	if err := os.MkdirAll(built, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(built, "widget.txt"), []byte("shop 1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := connect(t, ws)
	params := protocol.PublishParams{Project: protocol.Project{Name: "shop", Root: "apps/shop", Kind: protocol.ProjectApp}, Version: "1.0.0", DryRun: true}
	var res protocol.PublishResult

	if err := c.Call(protocol.MethodPublish, params, &res, 0); err != nil || res.Published || !strings.Contains(res.Skipped, "dry run") {
		t.Fatalf("dry run: %+v %v", res, err)
	}
	if _, err := os.Stat(filepath.Join(ws, "dist", "deploy")); err == nil {
		t.Fatal("a dry run wrote the target")
	}
	params.DryRun = false
	if err := c.Call(protocol.MethodPublish, params, &res, 0); err != nil || !res.Published {
		t.Fatalf("first deploy: %+v %v", res, err)
	}
	if err := c.Call(protocol.MethodPublish, params, &res, 0); err != nil || res.Published || !strings.Contains(res.Skipped, "already deployed") {
		t.Fatalf("second deploy must be skipped with a reason: %+v %v", res, err)
	}
}

func TestPublishWithoutABuildSaysWhatToDo(t *testing.T) {
	params := protocol.PublishParams{Project: protocol.Project{Name: "shop", Root: "apps/shop", Kind: protocol.ProjectApp}, Version: "1.0.0"}
	err := connect(t, t.TempDir()).Call(protocol.MethodPublish, params, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "smith run build -p shop") {
		t.Fatalf("got %v", err)
	}
}

func TestDoctorRejectsATargetOutsideTheWorkspace(t *testing.T) {
	app := protocol.Project{Name: "shop", Root: "apps/shop", Kind: protocol.ProjectApp, Settings: map[string]map[string]any{pluginName: {"target": "../elsewhere"}}}
	var res protocol.DoctorResult
	if err := connect(t, t.TempDir()).Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{app}}, &res, 0); err != nil {
		t.Fatal(err)
	}
	if f := res.Findings[0]; f.OK || f.Remedy == "" {
		t.Fatalf("%+v", f)
	}
}
