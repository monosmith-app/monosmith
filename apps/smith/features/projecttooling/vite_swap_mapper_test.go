package projecttooling

import (
	"strings"
	"testing"
)

func TestViteConfigBuildsIntoSmithsFolderWithTheReactPlugin(t *testing.T) {
	got := ViteConfig(ViteSettings{OutDir: "../../dist/apps/web", Base: "/web/"})
	for _, want := range []string{"import react from '@vitejs/plugin-react'", "plugins: [react()]", "outDir:      '../../dist/apps/web'", "base:    '/web/'"} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "test:") || strings.Contains(got, "vitest") {
		t.Errorf("no tests, no Vitest:\n%s", got)
	}
}

func TestViteConfigRunsTestsInJsdomAndKeepsCreateReactAppHabits(t *testing.T) {
	got := ViteConfig(ViteSettings{OutDir: "dist", Tests: true, SetupFile: "src/setupTests.ts", JSXInJS: true, ProcessEnv: true})
	for _, want := range []string{`/// <reference types="vitest/config" />`, "environment: 'jsdom'", "globals:     true", "setupFiles:  ['src/setupTests.ts']", "loader: 'jsx'", "envPrefix: 'REACT_APP_'"} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
}

func TestViteIndexHTMLDropsPublicURLAndLoadsTheEntry(t *testing.T) {
	got := ViteIndexHTML("<html><head><link href=\"%PUBLIC_URL%/favicon.ico\"></head><body><div id=\"root\"></div></body></html>", "/src/index.tsx")
	if strings.Contains(got, "%PUBLIC_URL%") || !strings.Contains(got, `href="/favicon.ico"`) {
		t.Errorf("public URL: %s", got)
	}
	if !strings.Contains(got, `<script type="module" src="/src/index.tsx"></script>`+"\n</body>") {
		t.Errorf("entry: %s", got)
	}
}

func TestViteScriptReplacesReactScripts(t *testing.T) {
	cases := []struct {
		in    string
		tests bool
		want  string
		keep  bool
	}{
		{"react-scripts build", true, "vite build", true},
		{"react-scripts start", true, "vite", true},
		{"react-scripts test --watchAll=false", true, "vitest run", true},
		{"react-scripts test", false, "", false},
		{"react-scripts eject", true, "", false},
		{"node tools/x.js", true, "node tools/x.js", true},
	}
	for _, c := range cases {
		got, keep := ViteScript(c.in, c.tests)
		if got != c.want || keep != c.keep {
			t.Errorf("%q: %q %v, want %q %v", c.in, got, keep, c.want, c.keep)
		}
	}
}
