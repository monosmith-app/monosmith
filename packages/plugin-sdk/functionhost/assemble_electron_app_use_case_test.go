package functionhost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssembleElectronAppWritesTheWorkFolder(t *testing.T) {
	ws := t.TempDir()
	project := filepath.Join(ws, "apps", "desk")
	write(t, filepath.Join(project, "package.json"), `{"name":"desk","main":"dist/main/app.handler.js"}`)
	write(t, filepath.Join(project, "src", "renderer", "index.html"), "<html></html>")
	write(t, filepath.Join(ws, "node_modules", "electron", "package.json"), `{"version":"31.0.1"}`)
	app := filepath.Join(ws, "dist", "electron", "apps", "desk", "app")
	if err := AssembleElectronApp(project, app); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(app, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"electron": "31.0.1"`, `"forge": "./forge.config.cjs"`, `"main": "dist/main/app.handler.js"`} {
		if !strings.Contains(string(manifest), want) {
			t.Errorf("the manifest lacks %s:\n%s", want, manifest)
		}
	}
	for _, f := range []string{"forge.config.cjs", filepath.Join("dist", "renderer", "index.html")} {
		if _, err := os.Stat(filepath.Join(app, f)); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
}

func TestAssembleElectronAppNeedsElectronInstalled(t *testing.T) {
	project := t.TempDir()
	write(t, filepath.Join(project, "package.json"), `{}`)
	err := AssembleElectronApp(project, filepath.Join(project, "app"))
	if err == nil || !strings.Contains(err.Error(), "smith install") {
		t.Fatalf("a missing electron says how to install it: %v", err)
	}
}
