package convertplan

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// webApp is a Vite app whose Nx build lists a folder per environment but
// whose build:local script writes only Vite's default dist.
func webApp(t *testing.T, viteConfig string) (*nxworkspace.Workspace, Input) {
	t.Helper()
	app := project("app", "apps/app", "app", []string{"package.json", "vite.config.ts"}, map[string]nxworkspace.Target{
		"build": target(t, `{"executor": "nx:run-commands", "outputs": ["{projectRoot}/dist", "{projectRoot}/dist-dev", "{projectRoot}/dist-preprod", "{projectRoot}/dist-prod"], "options": {"command": "npm run build:local -w app"}}`),
	})
	app.Package = &nxworkspace.PackageJSON{Name: "app", Private: true, Scripts: map[string]string{
		"build":       "npm run build:local",
		"build:local": "npm run typegen && vite build",
		"typegen":     "chakra typegen src/theme.ts",
	}}
	files := []string{"apps/app/package.json", "apps/app/vite.config.ts"}
	read := func(f string) string {
		if f == "apps/app/vite.config.ts" {
			return viteConfig
		}
		return ""
	}
	return &nxworkspace.Workspace{Projects: []*nxworkspace.Project{app}}, Input{DirName: "acme", Files: files, Read: read}
}

func TestPlanOffersTheOutputsTheBuildWrites(t *testing.T) {
	w, in := webApp(t, "export default defineConfig({ plugins: [react()] })")
	plan := MakePlan(w, in)

	if got := plan.Override("app").Tasks["build"].Outputs; !slices.Equal(got, []string{"dist"}) {
		t.Fatalf("outputs %v, want the detected [dist]", got)
	}
	want := OutputChoice{Project: "app", Task: "build", Declared: []string{"dist", "dist-dev", "dist-preprod", "dist-prod"}, Detected: []string{"dist"}}
	if len(plan.OutputChoices) != 1 || !slices.Equal(plan.OutputChoices[0].Declared, want.Declared) || !slices.Equal(plan.OutputChoices[0].Detected, want.Detected) {
		t.Fatalf("choices %+v, want %+v", plan.OutputChoices, want)
	}
	decide := ""
	for _, f := range plan.Findings {
		if f.Decide && f.Project == "app" {
			decide = f.Message
		}
	}
	if !strings.Contains(decide, "Nx declared dist, dist-dev, dist-preprod, dist-prod") || !strings.Contains(decide, "Suggested: cache dist") {
		t.Fatalf("decision %q", decide)
	}

	plan.KeepDeclaredOutputs(0)
	if got := plan.Override("app").Tasks["build"].Outputs; !slices.Equal(got, want.Declared) || !plan.OutputChoices[0].Kept {
		t.Fatalf("kept outputs %v", got)
	}
}

func TestPlanKeepsDeclaredOutputsWhenTheConfigNamesTheFolder(t *testing.T) {
	w, in := webApp(t, "export default defineConfig({ build: { outDir: 'www' } })")
	plan := MakePlan(w, in)

	if len(plan.OutputChoices) != 0 {
		t.Fatalf("choices %+v: the Vite config names the folder, so nothing is known to be wrong", plan.OutputChoices)
	}
	if got := plan.Override("app").Tasks["build"].Outputs; len(got) != 4 {
		t.Fatalf("outputs %v, want the four declared ones", got)
	}
}
