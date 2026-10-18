package projectgraph

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestADependsOnOverrideKeepsATaskThePluginAddedForIt(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "core", Root: "packages/core", Tasks: map[string]protocol.Task{
		"build": {Command: []string{"tsc"}},
	}}}
	merge(p, protocol.Contribution{Replace: map[string]protocol.Task{
		"e2e":          {Command: []string{"jest", "-c", "e2e"}, DependsOn: []string{"e2e-browsers", "^build", "build"}},
		"e2e-browsers": {Command: []string{"npx", "playwright", "install"}},
	}}, "plugin playwright (convention playwright)")
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"core": {Tasks: map[string]workspace.TaskOverride{
		"e2e": {DependsOn: []string{"^build"}},
	}}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	got := p.Tasks["e2e"].DependsOn
	if strings.Join(got, ",") != "^build,e2e-browsers" {
		t.Errorf("depends_on = %v; the override's list plus e2e-browsers, which only the plugin's e2e needs", got)
	}
	if src := p.Sources["e2e.dependsOn"]; !strings.Contains(src, "[projects.core.tasks.e2e]") || !strings.Contains(src, "e2e-browsers from plugin playwright") {
		t.Errorf("source = %q", src)
	}
	if d := p.Dropped["e2e"]; len(d) != 1 || d[0] != "build (from plugin playwright (convention playwright))" {
		t.Errorf("dropped = %v; build is the user's to leave out, and doctor names it", d)
	}
	if len(cfg.Projects["core"].Tasks["e2e"].DependsOn) != 1 {
		t.Error("the override's own list was changed")
	}
}

// The path OpenCraw takes: [conventions.playwright] replaces the e2e task the
// node plugin discovered from package.json, and adds e2e-browsers; the
// converted smith.toml then overrides e2e's depends_on.
func TestADependsOnOverrideKeepsWhatAConventionAddedForTheTask(t *testing.T) {
	for _, override := range [][]string{{"build"}, {"build", "^build"}, {"build", "core:build"}} {
		p := &Project{Project: protocol.Project{Name: "office-reader", Root: "packages/office-reader", Tasks: map[string]protocol.Task{
			"build": {Command: []string{"npm", "run", "build"}},
			"e2e":   {Command: []string{"npm", "run", "e2e"}},
		}}, Plugin: "node"}
		merge(p, protocol.Contribution{Replace: map[string]protocol.Task{
			"e2e":          {Command: []string{"npm", "run", "e2e"}, DependsOn: []string{"build", "e2e-browsers"}},
			"e2e-browsers": {Command: []string{"npx", "--no-install", "playwright", "install", "chromium"}},
		}}, "convention playwright")
		cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"office-reader": {Tasks: map[string]workspace.TaskOverride{
			"e2e": {DependsOn: override},
		}}}}
		if err := ApplyOverrides(p, cfg); err != nil {
			t.Fatal(err)
		}
		want := strings.Join(append(slices.Clone(override), "e2e-browsers"), ",")
		if got := strings.Join(p.Tasks["e2e"].DependsOn, ","); got != want {
			t.Errorf("override %v: depends_on = %s, want %s", override, got, want)
		}
		if src := p.Sources["e2e.dependsOn"]; !strings.Contains(src, "e2e-browsers from convention playwright") {
			t.Errorf("override %v: source = %q", override, src)
		}
	}
}

func TestAnImageCanWaitForTheAppsOwnPruneTask(t *testing.T) {
	// An app whose deploy output a hand-written script makes: the image
	// waits for it, and doctor names the node plugin's prune it left out.
	p := &Project{Project: protocol.Project{Name: "host", Root: "apps/host", Tasks: map[string]protocol.Task{
		"build":  {Command: []string{"npm", "run", "build"}},
		"prune":  {Command: []string{"{smith}", "npm-prune", "apps/host"}},
		"bundle": {Command: []string{"npm", "run", "bundle"}},
	}}, Plugin: "node"}
	merge(p, protocol.Contribution{Tasks: map[string]protocol.Task{
		"image": {Command: []string{"docker", "build", "."}, DependsOn: []string{"build", "prune"}},
	}}, "convention docker")
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"host": {Tasks: map[string]workspace.TaskOverride{
		"image": {DependsOn: []string{"build", "bundle"}},
	}}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(p.Tasks["image"].DependsOn, ","); got != "build,bundle" {
		t.Errorf("depends_on = %s", got)
	}
	if d := p.Dropped["image"]; len(d) != 1 || d[0] != "prune (from convention docker)" {
		t.Errorf("dropped = %v", d)
	}
}

func TestADependsOnOverrideRecordsWhatItDropsFromAConvention(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build": {Command: []string{"tsc"}},
		"lint":  {Command: []string{"eslint", "."}},
	}}, Plugin: "node"}
	merge(p, protocol.Contribution{DependsOn: map[string][]string{"lint": {"slice-check"}}}, "convention vertical-slices")
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"web": {Tasks: map[string]workspace.TaskOverride{
		"lint": {DependsOn: []string{"build"}},
	}}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if d := p.Dropped["lint"]; len(d) != 1 || d[0] != "slice-check (from convention vertical-slices)" {
		t.Errorf("dropped = %v", d)
	}
}

func TestADependsOnOverrideOfALanguagePluginsTaskIsRecorded(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"test": {Command: []string{"jest"}, DependsOn: []string{"build"}},
		"lint": {Command: []string{"eslint", "."}},
	}}, Plugin: "node"}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"web": {Tasks: map[string]workspace.TaskOverride{
		"test": {DependsOn: []string{"lint"}},
		"lint": {DependsOn: []string{"^build"}},
	}}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if got := p.Tasks["test"].DependsOn; len(got) != 1 || got[0] != "lint" {
		t.Errorf("depends_on = %v: only a task a contribution added for this one is kept", got)
	}
	if d := p.Dropped["test"]; len(d) != 1 || d[0] != "build (from plugin node)" {
		t.Errorf("dropped = %v", d)
	}
	if _, ok := p.Dropped["lint"]; ok {
		t.Error("lint had no dependencies to drop")
	}
}

func TestApplyOverridesSplitRepo(t *testing.T) {
	project := func() *Project {
		return &Project{Project: protocol.Project{Name: "client", Root: "packages/client", Tasks: map[string]protocol.Task{}}, Sources: map[string]string{}}
	}
	p := project()
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"client": {SplitRepo: "acme/client-php"}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if p.SplitRepo != "acme/client-php" || p.Release != nil {
		t.Errorf("split_repo is a project setting, not a release section (which builds assets): %+v", p)
	}

	cfg.Projects["client"] = workspace.ProjectOverride{SplitRepo: "https://github.com/acme/client-php"}
	err := ApplyOverrides(project(), cfg)
	if err == nil || !strings.Contains(err.Error(), "[projects.client]") || !strings.Contains(err.Error(), "owner/repo") {
		t.Errorf("a URL is refused, naming where it was set: %v", err)
	}
}

func TestAnAppsPackageTaskRunsInVerifyUnlessTurnedOff(t *testing.T) {
	task := map[string]protocol.Task{"package": {Command: []string{"npm", "pack"}}}
	app := &Project{Project: protocol.Project{Name: "api", Kind: protocol.ProjectApp, Tasks: map[string]protocol.Task{"package": task["package"]}}}
	lib := &Project{Project: protocol.Project{Name: "kit", Kind: protocol.ProjectPackage, Tasks: map[string]protocol.Task{"package": task["package"]}}}
	if err := ApplyOverrides(app, workspace.Config{}); err != nil || !app.Tasks["package"].Verify {
		t.Errorf("an app's package task is not marked for verify: %v", err)
	}
	if err := ApplyOverrides(lib, workspace.Config{}); err != nil || lib.Tasks["package"].Verify {
		t.Errorf("a package's package task is marked for verify: %v", err)
	}
	off := false
	app = &Project{Project: protocol.Project{Name: "api", Kind: protocol.ProjectApp, Tasks: map[string]protocol.Task{"package": task["package"]}}}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"api": {Tasks: map[string]workspace.TaskOverride{"package": {Verify: &off}}}}}
	if err := ApplyOverrides(app, cfg); err != nil || app.Tasks["package"].Verify {
		t.Errorf("verify = false did not turn it off: %v", err)
	}
}
