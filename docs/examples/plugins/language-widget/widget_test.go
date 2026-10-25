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

// connect serves the plugin in this process over pipes and runs the
// handshake, the way smith would over the plugin's stdin and stdout.
func connect(t *testing.T, workspace string) *host.Client {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(&widget{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: workspace})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// The conformance suite is what `smith plugin check --generate` runs.
func TestConforms(t *testing.T) {
	ws := t.TempDir()
	report := conformance.Check(connect(t, ws), conformance.Options{Workspace: ws})
	if !report.OK() {
		t.Fatal(report.Failures())
	}
}

func write(t *testing.T, ws, rel, content string) {
	t.Helper()
	full := filepath.Join(ws, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverTurnsUsesIntoImports(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "libs/core/widget.json", `{"name":"core","version":"1.0.0"}`)
	write(t, ws, "apps/shop/widget.json", `{"name":"shop","version":"1.0.0","kind":"app","uses":["core"]}`)
	var res protocol.DiscoverResult
	err := connect(t, ws).Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{"libs/core/widget.json", "apps/shop/widget.json"}}, &res, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("want 2 projects, got %+v", res.Projects)
	}
	core, shop := res.Projects[0], res.Projects[1]
	if core.Kind != protocol.ProjectLib || shop.Kind != protocol.ProjectApp {
		t.Errorf("kinds: %s, %s", core.Kind, shop.Kind)
	}
	if len(shop.Imports) != 1 || shop.Imports[0] != "core" || core.Exports[0] != "core" {
		t.Errorf("the graph edge shop -> core needs shop.imports=[core] and core.exports=[core]: %+v %+v", shop, core)
	}
	if got := core.Tasks["build"].Outputs[0]; got != "../../dist/libs/core" {
		t.Errorf("build output %q is not under dist/", got)
	}
}

func TestWriteVersionDryRunChangesNothing(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "libs/core/widget.json", "{\n  \"name\": \"core\",\n  \"version\": \"1.0.0\"\n}\n")
	c := connect(t, ws)
	project := protocol.Project{Name: "core", Root: "libs/core"}
	var res protocol.VersionResult
	if err := c.Call(protocol.MethodVersion, protocol.VersionParams{Project: project, Version: "1.1.0", DryRun: true}, &res, 0); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(ws, "libs", "core", "widget.json"))
	if !strings.Contains(string(raw), `"1.0.0"`) || len(res.Files) != 1 {
		t.Fatalf("a dry run must list the file and leave it alone: %v\n%s", res.Files, raw)
	}
	if err := c.Call(protocol.MethodVersion, protocol.VersionParams{Project: project, Version: "1.1.0"}, &res, 0); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(filepath.Join(ws, "libs", "core", "widget.json"))
	if !strings.Contains(string(raw), `"version": "1.1.0"`) {
		t.Fatalf("version not written:\n%s", raw)
	}
}

func TestDoctorOffersAFixForAMissingLicense(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "libs/core/widget.json", `{"name":"core","version":"1.0.0"}`)
	var res protocol.DoctorResult
	err := connect(t, ws).Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{{Name: "core", Root: "libs/core"}}}, &res, 0)
	if err != nil {
		t.Fatal(err)
	}
	f := res.Findings[0]
	if f.OK || f.Remedy == "" || f.Fix == nil || f.Fix.Safe || !strings.Contains(f.Fix.Files[0].Content, `"license": "MIT"`) {
		t.Fatalf("unexpected finding %+v", f)
	}
}
