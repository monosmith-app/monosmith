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
	go func() { _ = serve.Serve(&lint{}, pluginR, pluginW); _ = pluginW.Close() }()
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

func TestContributesALintTaskFromSettings(t *testing.T) {
	c := connect(t, t.TempDir())
	if got := c.Info.Provides; len(got) != 1 || got[0] != "lint" {
		t.Fatalf("provides %v", got)
	}
	strict := protocol.Project{Name: "strict", Root: "libs/strict", Settings: map[string]map[string]any{
		pluginName: {"require": []any{"description", "license", "version"}},
	}}
	plain := protocol.Project{Name: "plain", Root: "libs/plain"}
	var res protocol.ContributeResult
	if err := c.Call(protocol.MethodContribute, protocol.ContributeParams{Projects: []protocol.Project{strict, plain}}, &res, 0); err != nil {
		t.Fatal(err)
	}
	if got := res.Projects["strict"].Tasks["lint"].Command[2]; got != "description,license,version" {
		t.Errorf("strict requires %q", got)
	}
	if got := res.Projects["plain"].Tasks["lint"].Command[2]; got != "description,license" {
		t.Errorf("plain requires %q", got)
	}
}

func TestGenerateNeverOverwrites(t *testing.T) {
	ws := t.TempDir()
	c := connect(t, ws)
	params := protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core"}
	if err := os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755); err != nil {
		t.Fatal(err)
	}
	var res protocol.GenerateResult
	if err := c.Call(protocol.MethodGenerate, params, &res, 0); err != nil || len(res.Files) != 1 {
		t.Fatalf("first run: %v %v", res, err)
	}
	if err := c.Call(protocol.MethodGenerate, params, &res, 0); err != nil || len(res.Files) != 0 {
		t.Fatalf("second run must write nothing: %v %v", res, err)
	}
}

func TestLintNamesTheEmptyFields(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "widget.json"), []byte(`{"name":"x","description":""}`), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	err := runLint([]string{"lint", "description,license"})
	if err == nil || !strings.Contains(err.Error(), "description, license") {
		t.Fatalf("got %v", err)
	}
}
