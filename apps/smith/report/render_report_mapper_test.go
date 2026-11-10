package report

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/graphview"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// smallWorkspace is a workspace of three projects with one of each finding.
func smallWorkspace(t *testing.T) Data {
	t.Helper()
	g, err := projectgraph.New([]*projectgraph.Project{
		{Project: protocol.Project{Name: "web", Root: "apps/web", Kind: protocol.ProjectApp, Language: "node", Stack: "react-vitest"}, Deps: []string{"ui"}},
		{Project: protocol.Project{Name: "ui", Root: "libs/ui", Kind: protocol.ProjectPackage, Language: "node"}},
		{Project: protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp, Language: "go", Stack: "go"}, Deps: []string{"ui"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return Data{
		Workspace: "acme",
		Branch:    "main",
		GraphSVG:  graphview.SVG(g, nil),
		Projects: []Project{
			{Name: "api", Root: "apps/api", Kind: "app", Language: "go", Stack: "go", Conventions: []string{"vertical-slices"}, Deps: []string{"ui"}},
			{Name: "ui", Root: "libs/ui", Kind: "package", Language: "node"},
			{Name: "web", Root: "apps/web", Kind: "app", Language: "node", Stack: "react-vitest", Choices: []string{"react", "vitest"}, Deps: []string{"ui"}},
		},
		Health: Health{Passed: 2, Warned: 1, Failed: 1, Checks: []Check{
			{Name: "lockfile in step", Status: StatusFail, Detail: "web installs without a lockfile", Remedy: "run smith install"},
			{Name: "agent files", Status: StatusWarn, Detail: "AGENTS.md is missing", Remedy: "smith agent file"},
			{Name: "plugins trusted", Status: StatusPass},
			{Name: "slice names", Status: StatusPass, Detail: "3 projects checked"},
		}},
		Drift: Drift{Enabled: true, Current: 4, Items: []DriftItem{
			{Path: "apps/web/tsconfig.lib.json", Renderer: "node.tsconfig.references", Status: "drift", Owner: "smith", Detail: "someone edited the derived keys"},
		}},
		Release: Release{
			Items:     []ReleaseItem{{Project: "ui", Current: "1.2.0", Next: "1.3.0", Level: "minor", Reason: "2 releasing commit(s)"}},
			Unchanged: []string{"api"},
		},
	}
}

func TestRenderMatchesTheGoldenPage(t *testing.T) {
	got, err := Render(smallWorkspace(t))
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "small_workspace.golden.html")
	if *update {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("the page differs from %s; run go test ./report -update and read the diff", golden)
	}
}

func TestThePageIsSelfContained(t *testing.T) {
	got, err := Render(smallWorkspace(t))
	if err != nil {
		t.Fatal(err)
	}
	// The SVG namespace and the local marker reference are not requests.
	local := strings.NewReplacer("http://www.w3.org/2000/svg", "", "url(#arrow)", "").Replace(got)
	for _, banned := range []string{"<script", "<link", "<img", "@import", "src=", "url(", "http://", "https://"} {
		if strings.Contains(local, banned) {
			t.Errorf("the page contains %q: it must make no request", banned)
		}
	}
	for _, need := range []string{"prefers-color-scheme: dark", `lang="en"`, `role="img"`, "Skip to the report", "<caption>"} {
		if !strings.Contains(got, need) {
			t.Errorf("the page lacks %q", need)
		}
	}
}

func TestEveryStatusIsAWordNotOnlyAColour(t *testing.T) {
	got, _ := Render(smallWorkspace(t))
	for _, word := range []string{">Fail<", ">Warning<", ">Pass<"} {
		if !strings.Contains(got, word) {
			t.Errorf("the page lacks the status word %s", word)
		}
	}
}

func TestAnEmptyWorkspaceStillRenders(t *testing.T) {
	got, err := Render(Data{Workspace: "empty", Branch: "main"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Derived files are off", "Nothing would release", "Every check passed"} {
		if !strings.Contains(got, want) {
			t.Errorf("the page lacks %q", want)
		}
	}
}

func TestNamesAreEscaped(t *testing.T) {
	got, _ := Render(Data{Workspace: "<b>x</b>", Branch: "main"})
	if strings.Contains(got, "<b>x</b>") {
		t.Error("a workspace name was not escaped")
	}
}
