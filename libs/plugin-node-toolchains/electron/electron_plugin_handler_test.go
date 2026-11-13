package electron

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func describe(t *testing.T) protocol.InitializeResult {
	t.Helper()
	res, err := New("test").Describe(protocol.InitializeParams{})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestTheElectronChoiceIsOnTheShellLayerAndNeedsLintAndTest(t *testing.T) {
	choices := describe(t).Choices
	if len(choices) != 1 {
		t.Fatalf("one choice: %+v", choices)
	}
	c := choices[0]
	if err := protocol.ValidateChoice(c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "electron" || c.Axis != "framework" || c.Layer != "shell" || c.Title != "Electron" || !c.Plays(protocol.RoleScaffold) {
		t.Fatalf("%+v", c)
	}
	for _, want := range []string{"azure-functions", "aws-lambda", "gcp-functions", "tauri", "jest"} {
		found := false
		for _, x := range c.Conflicts {
			found = found || x == want
		}
		if !found {
			t.Fatalf("electron must conflict with %s: %v", want, c.Conflicts)
		}
	}
	if len(c.Aliases) == 0 || c.Aliases[0] != "desktop" {
		t.Fatalf("a search for desktop finds it: %v", c.Aliases)
	}
}

func TestThePlanHasTheMainProcessThePreloadAndThePage(t *testing.T) {
	plan, err := electronChoices.Plan([]string{"electron"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"src/main/app.handler.ts", "src/preload/bridge.handler.cts", "src/renderer/index.html", "package.json"} {
		if _, ok := plan.Files[path]; !ok {
			t.Fatalf("no %s in %v", path, plan.Files)
		}
	}
	if !strings.Contains(plan.Files["src/preload/bridge.handler.cts"], "contextBridge.exposeInMainWorld") {
		t.Fatal("the preload script uses contextBridge")
	}
	if !strings.Contains(plan.Files["src/renderer/index.html"], "Content-Security-Policy") {
		t.Fatal("the page has a content security policy")
	}
	if plan.Tasks["typecheck"] == "" {
		t.Fatalf("no typecheck task: %v", plan.Tasks)
	}
	if _, ok := plan.DevDependencies["electron"]; !ok {
		t.Fatalf("electron is a dev dependency: %v", plan.DevDependencies)
	}
}

// The app that start runs and Forge's out/ are in the work folder
// dist/electron/<project dir>; nothing of Forge's lands in dist/<project dir>
// (ADR 0092, owner decision 2).
func TestTheBuildAndThePackageUseTheWorkFolder(t *testing.T) {
	plan, err := electronChoices.Plan([]string{"electron"})
	if err != nil {
		t.Fatal(err)
	}
	if body := plan.Files["tsconfig.build.json"]; !strings.Contains(body, "[[.Up]]dist/electron/[[.Dir]]/app") || strings.Contains(body, "[[.Build]]") {
		t.Errorf("tsconfig.build.json does not use the work folder:\n%s", body)
	}
}

// No Electron app holds a copy of a smith script (decision D3): its tasks run
// smith's assemble command and Electron's own tools in the work folder.
func TestGenerateWritesNoScriptAndTheTasksRunSmithAndElectron(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("test")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	res, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "desk", Directory: "apps/desk", Stack: "s", Choices: []string{"electron", "vitest"}, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range res.Files {
		if strings.HasSuffix(f, ".mjs") || strings.Contains(f, "scripts/") {
			t.Errorf("an Electron app holds no copied script: %s", f)
		}
	}
	raw, err := os.ReadFile(filepath.Join(ws, "apps", "desk", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ Scripts map[string]string }
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for task, want := range map[string]string{
		"build":   "smith electron assemble",
		"start":   "cd ../../dist/electron/apps/desk/app && npx electron .",
		"package": "cd ../../dist/electron/apps/desk/app && npx electron-forge package",
	} {
		if !strings.Contains(manifest.Scripts[task], want) || strings.Contains(manifest.Scripts[task], "scripts/") {
			t.Errorf("%s is %q, want it to run %q", task, manifest.Scripts[task], want)
		}
	}
}

func TestOnlyAnAppIsGenerated(t *testing.T) {
	_, err := New("test").Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "x"})
	if err == nil || !strings.Contains(err.Error(), "is an app") {
		t.Fatalf("%v", err)
	}
}
