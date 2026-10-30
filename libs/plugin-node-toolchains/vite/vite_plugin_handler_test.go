package vite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDescribe(t *testing.T) {
	res, err := New("1.0.0").Describe(protocol.InitializeParams{})
	if err != nil || len(res.Stacks) != 6 {
		t.Fatal(err, res.Stacks)
	}
	for _, s := range res.Stacks {
		want := "vite,eslint,vitest"
		if s.Name == "react-tauri" {
			want += ",tauri"
		}
		if s.Scaffold != Name || strings.Join(s.Toolchain, ",") != want || s.Description == "" {
			t.Errorf("stack %+v", s)
		}
	}
}

func TestGenerate(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for stack, want := range map[string]struct{ entry, typecheck, dep, devDep string }{
		"react":    {"src/main.tsx", "tsc -p tsconfig.json --noEmit", "react-dom", "@vitejs/plugin-react"},
		"react-p5": {"src/bouncing-ball/bouncing-ball.client.ts", "tsc -p tsconfig.json --noEmit", "@p5-wrapper/react", "@types/react"},
		"vue":      {"src/app/app.component.vue", "vue-tsc -p tsconfig.json --noEmit", "vue", "vue-tsc"},
		"svelte":   {"src/app/app.component.svelte", "svelte-check --tsconfig ./tsconfig.json --fail-on-warnings", "svelte", "svelte-check"},
		"html":     {"src/app/show-greeting.handler.ts", "tsc -p tsconfig.json --noEmit", "", "vite"},
	} {
		t.Run(stack, func(t *testing.T) {
			ws := t.TempDir()
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: stack, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
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
				t.Fatal(err)
			}
			if manifest.Name != "@acme/web" || manifest.Smith.Kind != "app" || manifest.Smith.Stack != stack || manifest.Scripts["typecheck"] != want.typecheck || manifest.Scripts["build"] != "vite build" {
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
				t.Fatalf("tsconfig is JSON: %v", err)
			}
			if index := string(read("apps/web/index.html")); !strings.Contains(index, `src="/src/`+apps[stack].Entry+`"`) {
				t.Errorf("index.html loads the entry:\n%s", index)
			}
		})
	}
}

func TestGenerateRefuses(t *testing.T) {
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: t.TempDir()})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "ui", Directory: "libs/ui", Stack: "react"}); err == nil || !strings.Contains(err.Error(), "smith add app ui --stack react") {
		t.Errorf("a lib is refused: %v", err)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "angular"}); err == nil {
		t.Error("an unknown stack is refused")
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
	for stack, wantOverride := range map[string]bool{"react": true, "react-p5": true, "vue": false, "svelte": false, "html": false} {
		t.Run(stack, func(t *testing.T) {
			ws := t.TempDir()
			if err := os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"name":"acme","workspaces":["apps/*"],"overrides":{"eslint-plugin-jsx-a11y":{"eslint":"$eslint"}}}`), 0o644); err != nil {
				t.Fatal(err)
			}
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: stack, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
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
