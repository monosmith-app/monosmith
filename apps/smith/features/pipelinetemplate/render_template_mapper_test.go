package pipelinetemplate

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/cipipeline"
)

func TestEveryTemplateRenders(t *testing.T) {
	m := cipipeline.NewModel(cipipeline.Contributions{}, "1.2.3", "main")
	app := App{Name: "web", Root: "apps/web", Output: "dist"}
	does := map[string]string{
		"e2e":                   "smith run e2e --parallel 1",
		"package-zip":           "smith ci pack",
		"deploy-pages":          "path: apps/web/dist",
		"deploy-azure-function": "apps/web",
	}
	for _, tmpl := range Catalog {
		for _, provider := range []string{"github", "azure"} {
			content, ok := Render(tmpl.Name, provider, m, app)
			supported := strings.Contains(strings.Join(tmpl.Providers, ","), provider)
			if ok != supported {
				t.Errorf("%s/%s: rendered %v, catalog says %v", tmpl.Name, provider, ok, supported)
				continue
			}
			if !ok {
				continue
			}
			for _, want := range []string{"Yours from here", "smith ci setup", "1.2.3", does[tmpl.Name]} {
				if !strings.Contains(content, want) {
					t.Errorf("%s/%s lacks %q:\n%s", tmpl.Name, provider, want, content)
				}
			}
			for i, line := range strings.Split(content, "\n") {
				if strings.Contains(line, "\t") {
					t.Errorf("%s/%s line %d has a tab, which YAML forbids", tmpl.Name, provider, i+1)
				}
				if indent := len(line) - len(strings.TrimLeft(line, " ")); indent%2 != 0 {
					t.Errorf("%s/%s line %d is indented by %d", tmpl.Name, provider, i+1, indent)
				}
			}
		}
	}
	if Path("e2e", "github") != ".github/workflows/e2e.yml" || Path("e2e", "azure") != "azure-pipelines/e2e.yml" {
		t.Fatal("paths")
	}
}

func TestTheE2ETemplatesCacheBrowsersAndRunOnPullRequests(t *testing.T) {
	m := cipipeline.NewModel(cipipeline.Contributions{}, "1.2.3", "main")
	gh, _ := Render("e2e", "github", m, App{})
	az, _ := Render("e2e", "azure", m, App{})
	for _, want := range []string{"  pull_request:\n", "path: ~/.cache/ms-playwright\n", "'**/package.json'"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub e2e lacks %q:\n%s", want, gh)
		}
	}
	for _, want := range []string{"pr:\n  branches:\n    include:\n      - main\n", "PLAYWRIGHT_BROWSERS_PATH: $(Pipeline.Workspace)/ms-playwright\n", "  - task: Cache@2\n", "      path: $(PLAYWRIGHT_BROWSERS_PATH)\n"} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure e2e lacks %q:\n%s", want, az)
		}
	}
	if strings.Contains(az, "pr: none") {
		t.Error("the Azure e2e pipeline runs on pull requests too")
	}
}
