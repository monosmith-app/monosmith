package convertplan

import (
	"strings"
	"testing"
)

func TestRewriteScript(t *testing.T) {
	names := map[string]string{"@acme/web": "web", "api": "api"}
	for in, want := range map[string]string{
		"nx run-many -t build":                                  "smith run build",
		"nx run-many -t lint,test --projects=@acme/web,api":     "smith run lint test -p web,api",
		"nx run-many --target=test --parallel=3":                "smith run test",
		"npx nx affected -t typecheck,test,build":               "smith run typecheck test build --affected",
		"mnci lint --affected && nx affected -t typecheck,test": "smith run lint --affected && smith run typecheck test --affected",
		"nx run api:serve":                                      "smith run serve -p api",
		"nx run @acme/web:build:production":                     "smith run build -p web",
		"nx build api":                                          "smith run build -p api",
		"nx graph":                                              "smith graph",
		"nx release --dry-run":                                  "smith release --dry-run",
		"husky":                                                 "husky",
		"eslint . --fix --cache":                                "eslint . --fix --cache",
	} {
		got, ok := RewriteScript(in, names)
		if !ok || got != want {
			t.Errorf("%q: got %q (%v), want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"nx migrate latest", "nx run unknown:build", "nx release publish", "nx run-many -t build --exclude=x"} {
		if got, ok := RewriteScript(in, names); ok {
			t.Errorf("%q must not convert, got %q", in, got)
		}
	}
}

// OpenCraw's root scripts name package.json projects without their npm scope,
// as Nx allows, and keep shell commands after the Nx part.
func TestRewriteScriptResolvesUnscopedNamesInCompoundScripts(t *testing.T) {
	names := map[string]string{"@opencraw/core": "core", "@opencraw/cli": "cli", "@opencraw/studio-desktop": "studio-desktop", "@opencraw/vscode": "vscode-extension"}
	for in, want := range map[string]string{
		"nx run core:e2e":               "smith run e2e -p core",
		"nx run studio-desktop:package": "smith run package -p studio-desktop",
		"nx run @opencraw/vscode:lint":  "smith run lint -p vscode-extension",
		"nx build cli":                  "smith run build -p cli",
		"nx run core:build && node docs/how-it-works/capture/capture.mjs --check": "smith run build -p core && node docs/how-it-works/capture/capture.mjs --check",
		`nx run-many -t build -p core cli && node --test "examples/*/*.test.mjs"`: `smith run build -p core,cli && node --test "examples/*/*.test.mjs"`,
		"nx run-many -t build --projects=core,@opencraw/cli":                      "smith run build -p core,cli",
	} {
		got, ok := RewriteScript(in, names)
		if !ok || got != want {
			t.Errorf("%q: got %q (%v), want %q", in, got, ok, want)
		}
	}
	ambiguous := map[string]string{"@a/core": "core-a", "@b/core": "core-b"}
	if got, ok := RewriteScript("nx run core:build", ambiguous); ok {
		t.Errorf("a bare name two scopes share must not convert, got %q", got)
	}
}

func TestScriptGapExplainsArgumentsPassedToATask(t *testing.T) {
	in := "nx run studio-desktop:package -- --dir"
	if _, ok := RewriteScript(in, map[string]string{"studio-desktop": "studio-desktop"}); ok {
		t.Fatalf("%q converted, but smith run passes no arguments on", in)
	}
	if gap := ScriptGap(in); gap == "" || !strings.Contains(gap, "-- --dir") {
		t.Errorf("gap = %q", gap)
	}
	if gap := ScriptGap("nx migrate latest"); gap != "" {
		t.Errorf("gap for nx migrate = %q", gap)
	}
}
