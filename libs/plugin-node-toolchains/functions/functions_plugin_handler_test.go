package functions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDescribePublishesTheHostsTheTriggersAndTheRuntimes(t *testing.T) {
	res, err := New("1.0.0").Describe(protocol.InitializeParams{})
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]protocol.Choice{}
	for _, c := range res.Choices {
		if err := protocol.ValidateChoice(c); err != nil {
			t.Error(err)
		}
		byName[c.Name] = c
	}
	for _, host := range []string{"azure-functions", "aws-lambda", "gcp-functions"} {
		c := byName[host]
		if c.Axis != "framework" || !c.Plays(protocol.RoleScaffold) || c.Language != "node" {
			t.Errorf("%s: %+v", host, c)
		}
		for _, axis := range []string{"trigger", "runtime", "lint", "test"} {
			found := false
			for _, n := range c.Needs {
				found = found || n == axis
			}
			if !found {
				t.Errorf("%s must need the %s axis: %v", host, axis, c.Needs)
			}
		}
	}
	for name, axis := range map[string]string{"http-trigger": "trigger", "timer-trigger": "trigger", "queue-trigger": "trigger", "blob-trigger": "trigger", "node24": "runtime", "node22": "runtime"} {
		if byName[name].Axis != axis {
			t.Errorf("%s is on axis %q, want %q", name, byName[name].Axis, axis)
		}
	}
	for _, c := range res.Choices {
		if c.Default && (c.Axis == "trigger" || c.Axis == "runtime") {
			t.Errorf("%s: the trigger and the runtime are asked, never defaulted", c.Name)
		}
	}
	if !byName["http-trigger"].Suggested || !byName["node24"].Suggested || byName["node22"].Suggested {
		t.Error("http-trigger and node24 (the current LTS) are the suggested answers")
	}
	if got := byName["timer-trigger"].Conflicts; len(got) != 1 || got[0] != "gcp-functions" {
		t.Errorf("Google has no timer trigger: %v", got)
	}
}

func TestGenerateWritesAnAppForEveryHostAndTrigger(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for _, host := range []string{"azure-functions", "aws-lambda", "gcp-functions"} {
		for _, trigger := range []string{"http-trigger", "timer-trigger", "queue-trigger", "blob-trigger"} {
			if host == "gcp-functions" && trigger == "timer-trigger" {
				continue
			}
			t.Run(host+"/"+trigger, func(t *testing.T) {
				ws := t.TempDir()
				p := New("1.0.0")
				_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
				res, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "api", Directory: "apps/api", Stack: "s", Choices: []string{host, trigger, "node24"}, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}})
				if err != nil {
					t.Fatal(err)
				}
				read := func(rel string) string {
					raw, err := os.ReadFile(filepath.Join(ws, "apps", "api", filepath.FromSlash(rel)))
					if err != nil {
						t.Fatalf("%s: %v (wrote %v)", rel, err, res.Files)
					}
					return string(raw)
				}
				var manifest struct {
					Name    string
					Main    string
					Engines map[string]string
					Scripts map[string]string
				}
				if err := json.Unmarshal([]byte(read("package.json")), &manifest); err != nil {
					t.Fatalf("package.json is JSON: %v", err)
				}
				if manifest.Name != "@acme/api" || manifest.Engines["node"] != ">=24" {
					t.Errorf("manifest %+v", manifest)
				}
				for _, task := range []string{"build", "typecheck", "start", "package"} {
					if manifest.Scripts[task] == "" {
						t.Errorf("task %s is missing: %v", task, manifest.Scripts)
					}
				}
				if !strings.Contains(manifest.Scripts["build"], "&&") {
					t.Errorf("a script is shell text, not HTML-escaped: %q", manifest.Scripts["build"])
				}
				if !strings.Contains(read("tsconfig.build.json"), `"outDir": "../../dist/apps/api/dist"`) {
					t.Errorf("the build goes to the workspace dist/:\n%s", read("tsconfig.build.json"))
				}
				read("src/greeting/greet.use-case.ts")
				read("src/functions/greeting.handler.ts")
				read("src/functions/greeting.handler.spec.ts")
				if _, err := os.Stat(filepath.Join(ws, "apps", "api", "local.settings.json")); err == nil {
					t.Error("local.settings.json holds secrets: only the sample is written")
				}
				switch host {
				case "azure-functions":
					var hostJSON, settings map[string]any
					if json.Unmarshal([]byte(read("host.json")), &hostJSON) != nil || json.Unmarshal([]byte(read("local.settings.sample.json")), &settings) != nil {
						t.Error("host.json and the settings sample are JSON")
					}
					if manifest.Main != "dist/functions/*.js" || !strings.Contains(read(".gitignore"), "local.settings.json") {
						t.Errorf("main %q, gitignore %q", manifest.Main, read(".gitignore"))
					}
				case "aws-lambda":
					template := read("template.yaml")
					for _, want := range []string{"Transform: AWS::Serverless-2016-10-31", "Runtime: nodejs24.x", "CodeUri: ../../dist/apps/api", "Handler: dist/functions/greeting.handler.handler"} {
						if !strings.Contains(template, want) {
							t.Errorf("template.yaml lacks %q:\n%s", want, template)
						}
					}
					read("events/greeting.json")
				case "gcp-functions":
					if manifest.Main != "dist/index.js" || !strings.Contains(read("src/index.ts"), "./functions/index.ts") {
						t.Errorf("the Functions Framework loads main: %q", manifest.Main)
					}
				}
			})
		}
	}
}

func TestGenerateRefuses(t *testing.T) {
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: t.TempDir()})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "x", Directory: "libs/x", Choices: []string{"azure-functions"}}); err == nil || !strings.Contains(err.Error(), "an app") {
		t.Errorf("a lib is refused: %v", err)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "x", Directory: "apps/x", Stack: "react", Choices: []string{"react"}}); err == nil || !strings.Contains(err.Error(), "azure-functions") {
		t.Errorf("a stack of another plugin is refused: %v", err)
	}
}
