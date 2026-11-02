package adoption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

func TestDetectedProvidersKnowsEveryProvider(t *testing.T) {
	exists := func(have ...string) func(string) bool {
		return func(p string) bool {
			for _, h := range have {
				if h == p {
					return true
				}
			}
			return false
		}
	}
	cases := []struct {
		have []string
		want string
	}{
		{nil, ""},
		{[]string{".gitlab-ci.yml"}, "gitlab"},
		{[]string{".circleci"}, "circleci"},
		{[]string{"bitbucket-pipelines.yml", ".github/workflows"}, "github,bitbucket"},
	}
	for _, c := range cases {
		if got := strings.Join(detectedProviders(exists(c.have...)), ","); got != c.want {
			t.Errorf("%v: got %q, want %q", c.have, got, c.want)
		}
	}
	if got := providersOf(exists()); len(got) != 1 || got[0] != "github" {
		t.Errorf("nothing detected falls back to github: %v", got)
	}
}

func TestGitignoreWithAddsOnlyWhatIsMissing(t *testing.T) {
	cases := []struct {
		name, before string
		node         bool
		want         string
	}{
		{"empty", "", false, "\n# smith\n.smith/\n"},
		{"already ignores both", ".smith/\nnode_modules/\n", true, ".smith/\nnode_modules/\n"},
		{"adds both, fixes the missing newline", "dist", true, "dist\n\n# smith\n.smith/\nnode_modules/\n"},
		{"node only when there are Node projects", "dist\n", false, "dist\n\n# smith\n.smith/\n"},
	}
	for _, c := range cases {
		if got := gitignoreWith(c.before, c.node); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestPlanInitShowsWhatMovesWhatIsCreatedAndWhatIsEditedWithADiff(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".github/workflows/ci.yml", "name: mine\n")
	write(".gitignore", "dist\n")
	write("package.json", "{\n  \"name\": \"acme\",\n  \"private\": true\n}\n")
	write(".mcp.json", "{\n  \"mcpServers\": {\n    \"mine\": {\"command\": \"x\"}\n  }\n}\n")
	r := Report{Root: root, Name: "acme", Projects: []ProjectLine{{Name: "web", Root: "apps/web", Language: "node"}}}
	c := adoptdecision.Choices{Gitattributes: adoptdecision.GitattributesAdd, Agent: adoptdecision.AgentAdd, CI: []string{"github"}, Existing: adoptdecision.ExistingReplace}
	plan := PlanInit(r, c, false)
	var out strings.Builder
	operationplan.Render(&out, plan)
	text := out.String()
	for _, want := range []string{
		"Pipelines of yours", ".github/workflows/ci.yml.pre-smith",
		"create  smith.toml", "create  .gitattributes", "create  .devcontainer/devcontainer.json",
		"change  package.json", "+  \"workspaces\"", "change  .gitignore", "+node_modules/",
		"change  .mcp.json",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the plan lacks %q:\n%s", want, text)
		}
	}
	// Skipping the agent and the line endings leaves both out of the plan.
	c.Agent, c.Gitattributes = adoptdecision.AgentSkip, adoptdecision.GitattributesSkip
	out.Reset()
	operationplan.Render(&out, PlanInit(r, c, false))
	if strings.Contains(out.String(), ".gitattributes") || strings.Contains(out.String(), ".mcp.json") {
		t.Errorf("a skipped choice appears in the plan:\n%s", out.String())
	}
}
