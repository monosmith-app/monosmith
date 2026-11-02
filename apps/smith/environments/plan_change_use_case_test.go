package environments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func webApp(settings webenv.Settings) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Settings: map[string]map[string]any{webenv.PluginName: settings.ToMap()}}}
}

func TestAddingAnEnvironmentCreatesItsFileAndScript(t *testing.T) {
	c, err := PlanChange(webApp(webenv.Settings{Names: webenv.DefaultNames}), []string{"qa"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(c.Created[".env.qa"], "never put a secret here") {
		t.Errorf("the file has the warning: %v", c.Created)
	}
	if c.ScriptsAdded["build:qa"] != "vite build --mode qa --outDir ../../dist/apps/web/qa" {
		t.Errorf("scripts %v", c.ScriptsAdded)
	}
	if strings.Join(c.After.Names, ",") != "lcl,dev,uat,prd,qa" {
		t.Errorf("names %v", c.After.Names)
	}
	text := ""
	for _, g := range c.Plan().Groups {
		for _, l := range g.Lines {
			text += l.Path + " " + l.Note + "\n"
		}
	}
	for _, want := range []string{"apps/web/.env.qa", "adds the script build:qa", "smith.toml"} {
		if !strings.Contains(text, want) {
			t.Errorf("the plan misses %q:\n%s", want, text)
		}
	}
}

func TestRemovingKeepsTheEnvFileAndRefusesTheDefault(t *testing.T) {
	app := webApp(webenv.Settings{Names: webenv.DefaultNames})
	c, err := PlanChange(app, nil, []string{"dev"})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.ScriptsRemoved) != 1 || c.ScriptsRemoved[0] != "build:dev" || len(c.Kept) != 1 || c.Kept[0] != "apps/web/.env.dev" {
		t.Errorf("got %+v", c)
	}
	for _, tc := range []struct {
		add, remove []string
		want        string
	}{
		{nil, []string{"prd"}, "default environment"},
		{nil, []string{"qa"}, "has no environment"},
		{[]string{"uat"}, nil, "already"},
		{[]string{"Qa"}, nil, "not an environment name"},
		{nil, nil, "name the environments"},
	} {
		if _, err := PlanChange(app, tc.add, tc.remove); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%v %v: got %v, want %q", tc.add, tc.remove, err, tc.want)
		}
	}
	plain := &projectgraph.Project{Project: protocol.Project{Name: "api", Root: "apps/api"}}
	if _, err := PlanChange(plain, []string{"qa"}, nil); err == nil || !strings.Contains(err.Error(), "no environments") {
		t.Errorf("an app with one build: %v", err)
	}
}

func TestApplyWritesFilesScriptsAndSettings(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "apps", "web")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := "{\n  \"name\": \"web\",\n  \"scripts\": {\n    \"build\": \"vite build --mode prd --outDir ../../dist/apps/web/prd\",\n    \"build:dev\": \"x\"\n  }\n}\n"
	toml := "[workspace]\nname = \"acme\"\n"
	for path, body := range map[string]string{filepath.Join(app, "package.json"): manifest, filepath.Join(root, "smith.toml"): toml, filepath.Join(app, ".env.qa"): "VITE_MINE=1\n"} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	settings := webenv.Settings{Names: []string{"dev", "prd"}}
	c, err := PlanChange(webApp(settings), []string{"qa"}, []string{"dev"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(root, c); err != nil {
		t.Fatal(err)
	}
	read := func(p string) string {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if got := read(filepath.Join(app, ".env.qa")); got != "VITE_MINE=1\n" {
		t.Errorf("a file that exists is never replaced: %q", got)
	}
	pkg := read(filepath.Join(app, "package.json"))
	if !strings.Contains(pkg, `"build:qa": "vite build --mode qa --outDir ../../dist/apps/web/qa"`) || strings.Contains(pkg, "build:dev") {
		t.Errorf("package.json:\n%s", pkg)
	}
	if got := read(filepath.Join(root, "smith.toml")); !strings.Contains(got, `environments = ["prd", "qa"]`) {
		t.Errorf("smith.toml:\n%s", got)
	}
}
