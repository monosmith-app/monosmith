package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/fs"
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
	go func() { _ = serve.Serve(&scaffold{}, pluginR, pluginW); _ = pluginW.Close() }()
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

func generate(t *testing.T, ws string, p protocol.GenerateParams) (protocol.GenerateResult, error) {
	t.Helper()
	var res protocol.GenerateResult
	err := connect(t, ws).Call(protocol.MethodGenerate, p, &res, 0)
	return res, err
}

func TestWritesTheServiceWithTheRecipesPort(t *testing.T) {
	ws := t.TempDir()
	res, err := generate(t, ws, protocol.GenerateParams{
		Kind: protocol.ProjectApp, Name: "shop", Directory: "apps/shop", Stack: stackName,
		Workspace: protocol.WorkspaceInfo{Name: "acme"}, Settings: map[string]any{"port": float64(9000)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Files) != 4 {
		t.Fatalf("files: %v", res.Files)
	}
	raw, err := os.ReadFile(filepath.Join(ws, "apps", "shop", "widget.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ Stack string }
	if err := json.Unmarshal(raw, &manifest); err != nil || manifest.Stack != stackName {
		t.Fatalf("the manifest must record the stack so discovery reports it back: %s (%v)", raw, err)
	}
	service, _ := os.ReadFile(filepath.Join(ws, "apps", "shop", "service.json"))
	if !strings.Contains(string(service), `"port": 9000`) {
		t.Fatalf("service.json: %s", service)
	}
}

func TestOnlyAnAppIsAService(t *testing.T) {
	_, err := generate(t, t.TempDir(), protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "x", Directory: "libs/x"})
	if err == nil || !strings.Contains(err.Error(), "smith add app x --stack widget-service") {
		t.Fatalf("got %v", err)
	}
}

func TestABadPortIsRefusedWithTheSettingNamed(t *testing.T) {
	_, err := generate(t, t.TempDir(), protocol.GenerateParams{
		Kind: protocol.ProjectApp, Name: "x", Directory: "apps/x", Settings: map[string]any{"port": "http"},
	})
	if err == nil || !strings.Contains(err.Error(), "settings.widget-service.port") {
		t.Fatalf("got %v", err)
	}
}

// update rewrites the golden files: go test -update
var update = flag.Bool("update", false, "rewrite testdata/golden")

// TestGolden compares everything the scaffold writes with the files in
// testdata/golden, so a change to a template shows up as a diff in review.
func TestGolden(t *testing.T) {
	ws := t.TempDir()
	_, err := generate(t, ws, protocol.GenerateParams{
		Kind: protocol.ProjectApp, Name: "shop", Directory: "apps/shop", Stack: stackName,
		Workspace: protocol.WorkspaceInfo{Name: "acme"}, Settings: map[string]any{"port": float64(9000)},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "golden")
	if *update {
		if err := os.RemoveAll(golden); err != nil {
			t.Fatal(err)
		}
	}
	var seen []string
	err = filepath.WalkDir(ws, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(ws, path)
		seen = append(seen, rel)
		got, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		want := filepath.Join(golden, rel+".golden")
		if *update {
			if err := os.MkdirAll(filepath.Dir(want), 0o755); err != nil {
				return err
			}
			return os.WriteFile(want, got, 0o644)
		}
		wantBytes, err := os.ReadFile(want)
		if err != nil {
			t.Errorf("%s has no golden file (run go test -update): %v", rel, err)
			return nil
		}
		if string(got) != string(wantBytes) {
			t.Errorf("%s differs from its golden file:\n--- got\n%s\n--- want\n%s", rel, got, wantBytes)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 4 {
		t.Errorf("the scaffold wrote %d files, the golden set has 4: %v", len(seen), seen)
	}
}
