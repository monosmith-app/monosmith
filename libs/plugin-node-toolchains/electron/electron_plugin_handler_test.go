package electron

import (
	"strings"
	"testing"

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
	for _, path := range []string{"src/main/app.handler.ts", "src/preload/bridge.handler.cts", "src/renderer/index.html", "package.json", "scripts/package.mjs"} {
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
	for _, task := range []string{"start", "build", "typecheck", "package"} {
		if plan.Tasks[task] == "" {
			t.Fatalf("no %s task: %v", task, plan.Tasks)
		}
	}
	if _, ok := plan.DevDependencies["electron"]; !ok {
		t.Fatalf("electron is a dev dependency: %v", plan.DevDependencies)
	}
}

func TestOnlyAnAppIsGenerated(t *testing.T) {
	_, err := New("test").Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "x"})
	if err == nil || !strings.Contains(err.Error(), "is an app") {
		t.Fatalf("%v", err)
	}
}
