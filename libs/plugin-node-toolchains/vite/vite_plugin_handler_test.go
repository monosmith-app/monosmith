package vite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDescribePublishesTheChoicesAndTheStacks(t *testing.T) {
	res, err := New("1.0.0").Describe(protocol.InitializeParams{})
	if err != nil || len(res.Stacks) != 6 {
		t.Fatal(err, res.Stacks)
	}
	for _, s := range res.Stacks {
		if len(s.Choices) == 0 || s.Description == "" || s.Language != "" {
			t.Errorf("a published stack is a saved selection, the host derives the rest: %+v", s)
		}
	}
	names := map[string]string{}
	for _, c := range res.Choices {
		if err := protocol.ValidateChoice(c); err != nil {
			t.Error(err)
		}
		names[c.Name] = c.Axis
	}
	for name, axis := range map[string]string{"react": "framework", "vue": "framework", "svelte": "framework", "html": "framework", "p5": "library", "bootstrap": "style"} {
		if names[name] != axis {
			t.Errorf("choice %s is on axis %q, want %q", name, names[name], axis)
		}
	}
}

func TestGenerate(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for name, want := range map[string]struct {
		choices                       []string
		entry, typecheck, dep, devDep string
		loads                         string
	}{
		"react":    {[]string{"react"}, "src/main.tsx", "tsc -p tsconfig.json --noEmit", "react-dom", "@vitejs/plugin-react", "main.tsx"},
		"react-p5": {[]string{"react", "p5"}, "src/bouncing-ball/bouncing-ball.client.ts", "tsc -p tsconfig.json --noEmit", "@p5-wrapper/react", "@types/react", "main.tsx"},
		"vue":      {[]string{"vue"}, "src/app/app.component.vue", "vue-tsc -p tsconfig.json --noEmit", "vue", "vue-tsc", "main.ts"},
		"svelte":   {[]string{"svelte"}, "src/app/app.component.svelte", "svelte-check --tsconfig ./tsconfig.json --fail-on-warnings", "svelte", "svelte-check", "main.ts"},
		"html":     {[]string{"html"}, "src/app/show-greeting.handler.ts", "tsc -p tsconfig.json --noEmit", "", "vite", "main.ts"},
	} {
		t.Run(name, func(t *testing.T) {
			ws := t.TempDir()
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: name, Choices: want.choices, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
				t.Fatal(err)
			}
			read := func(rel string) []byte {
				raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
				if err != nil {
					t.Fatal(err)
				}
				return raw
			}
			read("apps/web/" + want.entry)
			read("apps/web/src/greeting/greet.use-case.ts")
			var manifest struct {
				Name         string
				Smith        struct{ Kind, Stack string }
				Scripts      map[string]string
				Dependencies map[string]string
			}
			if err := json.Unmarshal(read("apps/web/package.json"), &manifest); err != nil {
				t.Fatalf("package.json is JSON: %v\n%s", err, read("apps/web/package.json"))
			}
			if manifest.Name != "@acme/web" || manifest.Smith.Kind != "app" || manifest.Smith.Stack != name || manifest.Scripts["typecheck"] != want.typecheck || manifest.Scripts["build"] != "vite build" || manifest.Scripts["start"] != "vite" {
				t.Errorf("manifest %+v", manifest)
			}
			if want.dep != "" && manifest.Dependencies[want.dep] == "" {
				t.Errorf("the app depends on %s: %+v", want.dep, manifest.Dependencies)
			}
			var root struct{ DevDependencies map[string]string }
			if err := json.Unmarshal(read("package.json"), &root); err != nil {
				t.Fatal(err)
			}
			if root.DevDependencies[want.devDep] == "" || root.DevDependencies["vite"] == "" {
				t.Errorf("the root has the tooling: %+v", root.DevDependencies)
			}
			if config := string(read("apps/web/vite.config.ts")); !strings.Contains(config, "outDir:      '../../dist/apps/web',") {
				t.Errorf("the build goes to dist/:\n%s", config)
			}
			var tsconfig map[string]any
			if err := json.Unmarshal(read("apps/web/tsconfig.json"), &tsconfig); err != nil {
				t.Fatalf("tsconfig is JSON: %v\n%s", err, read("apps/web/tsconfig.json"))
			}
			if index := string(read("apps/web/index.html")); !strings.Contains(index, `src="/src/`+want.loads+`"`) {
				t.Errorf("index.html loads the entry:\n%s", index)
			}
		})
	}
}

// A style joins any framework through the slots, with no line about the framework.
func TestAStyleJoinsEveryFramework(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for framework, entry := range map[string]string{"react": "main.tsx", "vue": "main.ts", "svelte": "main.ts", "html": "main.ts"} {
		t.Run(framework, func(t *testing.T) {
			ws := t.TempDir()
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "s", Choices: []string{framework, "bootstrap"}, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
				t.Fatal(err)
			}
			main, err := os.ReadFile(filepath.Join(ws, "apps", "web", "src", entry))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(main), "import 'bootstrap/dist/css/bootstrap.min.css'\n\n") {
				t.Errorf("the entry imports the stylesheet after the other imports:\n%s", main)
			}
			manifest, _ := os.ReadFile(filepath.Join(ws, "apps", "web", "package.json"))
			if !strings.Contains(string(manifest), `"bootstrap": "^5.`) {
				t.Errorf("the app depends on bootstrap:\n%s", manifest)
			}
		})
	}
}

func TestGenerateRefuses(t *testing.T) {
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: t.TempDir()})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "ui", Directory: "libs/ui", Stack: "react", Choices: []string{"react"}}); err == nil || !strings.Contains(err.Error(), "smith add app ui --stack react") {
		t.Errorf("a lib is refused: %v", err)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "angular"}); err == nil || !strings.Contains(err.Error(), "react") {
		t.Errorf("a stack with no choice of this plugin is refused, and the message names the choices: %v", err)
	}
}

// A React app next to a Jest project: npm cannot install while the optional
// peer chain of @vitejs/plugin-react reaches Babel 8 and Jest holds Babel 7
// (issue 584), so the root pins that chain to Babel 7 and keeps the overrides
// another toolchain wrote.
func TestReactPinsTheBabelPeerChainBesideTheWorkspacesOwnOverrides(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for _, choices := range [][]string{{"react"}, {"react", "p5"}, {"vue"}, {"svelte"}, {"html"}} {
		wantOverride := slices.Contains(choices, "react")
		t.Run(strings.Join(choices, "+"), func(t *testing.T) {
			ws := t.TempDir()
			if err := os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name":"acme","workspaces":["apps/*"],"overrides":{"eslint-plugin-jsx-a11y":{"eslint":"$eslint"}}}`), 0o644); err != nil {
				t.Fatal(err)
			}
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "s", Choices: choices, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(ws, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			var root struct{ Overrides map[string]json.RawMessage }
			if err := json.Unmarshal(raw, &root); err != nil {
				t.Fatal(err)
			}
			if _, ok := root.Overrides["eslint-plugin-jsx-a11y"]; !ok {
				t.Errorf("the workspace's own override was lost:\n%s", raw)
			}
			if got := string(root.Overrides["@babel/plugin-transform-runtime"]); wantOverride != (got == `"^7.29.0"`) {
				t.Errorf("Babel override %q, want it %v:\n%s", got, wantOverride, raw)
			}
		})
	}
}
