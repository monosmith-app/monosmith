package recipedrift

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// fake is a plugin that only describes itself.
type fake struct{ info protocol.InitializeResult }

func (f fake) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return f.info, nil
}

func connect(t *testing.T, name string, kind protocol.PluginKind, roles []protocol.Role, languages []string, provides ...string) *host.Client {
	t.Helper()
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	p := fake{protocol.InitializeResult{Plugin: protocol.Peer{Name: name}, Kind: kind, Roles: roles, Languages: languages, Provides: provides}}
	go func() { _ = serve.Serve(p, toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, host.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// plugins are python, and the toolchain plugins ruff and flake8 (both lint),
// pytest (test) and vertical-slices (structure).
func plugins(t *testing.T) *pluginhost.PluginSet {
	tc := []protocol.Role{protocol.RoleToolchain}
	return &pluginhost.PluginSet{Clients: []*host.Client{
		connect(t, "python", protocol.KindLanguage, nil, []string{"python"}),
		connect(t, "ruff", protocol.KindConvention, tc, []string{"python"}, "lint"),
		connect(t, "flake8", protocol.KindConvention, tc, []string{"python"}, "lint"),
		connect(t, "pytest", protocol.KindConvention, tc, []string{"python"}, "test"),
	}}
}

const teamRecipe = `name = "team"
version = "%s"

[choices.language]
options = ["python"]
%s
[when.language.python.use]
language = "python"
lint = "%s"
test = "pytest"

[when.language.python.settings.ruff]
line = %d
`

func recipeAt(version, extraChoices, lint string, line int) string {
	return fmt.Sprintf(teamRecipe, version, extraChoices, lint, line)
}

// A project made from team 1.2.0 with ruff at line 100.
func sessionFor(t *testing.T, recipeText string) (*session.Session, *projectgraph.Project, workspace.ProjectRecipe) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "team.recipe.toml"), []byte(recipeText), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := workspace.ProjectRecipe{Name: "team", Version: "1.2.0", Source: "team.recipe.toml", Answers: map[string]string{"language": "python"}}
	cfg := workspace.Config{
		Workspace: workspace.WorkspaceSection{Name: "acme", Recipe: "team.recipe.toml"},
		Stacks: map[string]workspace.StackConfig{
			"team-python": {Language: "python", Toolchain: []string{"ruff", "pytest"}},
		},
		Projects: map[string]workspace.ProjectOverride{
			"orders": {Recipe: &rec, Settings: map[string]map[string]any{"ruff": {"line": int64(100)}}},
		},
	}
	set := plugins(t)
	stacks, err := pluginhost.NewCatalog(set, cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := &session.Session{Root: root, Config: cfg, Plugins: set, Stacks: stacks, Version: "1.0.0"}
	p := &projectgraph.Project{Project: protocol.Project{Name: "orders", Kind: protocol.ProjectPackage, Stack: "team-python"}}
	return s, p, rec
}

func check(t *testing.T, s *session.Session, p *projectgraph.Project, rec workspace.ProjectRecipe) Report {
	t.Helper()
	return CheckProject(s, &recipe.Loader{Root: s.Root}, p, rec)
}

func TestCheckProjectFollowsItsRecipe(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.2.0", "", "ruff", 100))
	r := check(t, s, p, rec)
	if r.Drifted() || r.Problem != "" {
		t.Fatalf("an unchanged recipe is not drift: %+v", r)
	}
	if r.Summary() != "orders follows its recipe team 1.2.0" {
		t.Errorf("summary %q", r.Summary())
	}
}

func TestCheckProjectSeesASettingChangedInANewVersion(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.3.0", "", "ruff", 88))
	r := check(t, s, p, rec)
	if !r.Drifted() {
		t.Fatalf("1.3.0 changes the lint settings: %+v", r)
	}
	want := "orders is on team 1.2.0; the recipe is 1.3.0, which changes the settings of the lint plugin ruff (line: 100 → 88)"
	if r.Summary() != want {
		t.Errorf("summary\n got %s\nwant %s", r.Summary(), want)
	}
	if r.Settings["ruff"]["line"] != int64(88) || r.Stack.Name != "team-python" {
		t.Errorf("the report must carry what the project becomes: %+v %+v", r.Settings, r.Stack)
	}
}

func TestCheckProjectSeesAPluginSwap(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.3.0", "", "flake8", 100))
	r := check(t, s, p, rec)
	if !strings.Contains(r.Summary(), "adds the lint plugin flake8") || !strings.Contains(r.Summary(), "drops the toolchain plugin ruff") {
		t.Errorf("summary %q", r.Summary())
	}
	if len(r.Added) != 1 || r.Added[0] != "flake8" {
		t.Errorf("added %v", r.Added)
	}
}

func TestCheckProjectRecordsANewChoiceWithADefault(t *testing.T) {
	extra := "\n[choices.kind]\ndefault = \"library\"\noptions = [\"library\", \"service\"]\n"
	s, p, rec := sessionFor(t, recipeAt("1.2.0", extra, "ruff", 100))
	r := check(t, s, p, rec)
	if !r.Drifted() || !strings.Contains(r.Summary(), `asks a new choice, kind, which takes its default "library"`) {
		t.Errorf("a new choice with a default: %s", r.Summary())
	}
	if r.Answers["kind"] != "library" {
		t.Errorf("answers %v", r.Answers)
	}
}

func TestCheckProjectNamesANewChoiceWithNoAnswer(t *testing.T) {
	extra := "\n[choices.kind]\noptions = [\"library\", \"service\"]\n"
	s, p, rec := sessionFor(t, recipeAt("1.3.0", extra, "ruff", 100))
	r := check(t, s, p, rec)
	if r.Drifted() || !strings.Contains(r.Problem, `"kind"`) || !strings.Contains(r.Problem, "[projects.orders.recipe.answers]") {
		t.Errorf("a choice with no answer is a problem that names it: %+v", r)
	}
}

func TestCheckProjectReportsARecipeThatCannotBeLoaded(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.2.0", "", "ruff", 100))
	if err := os.Remove(filepath.Join(s.Root, "team.recipe.toml")); err != nil {
		t.Fatal(err)
	}
	r := check(t, s, p, rec)
	if r.Drifted() || !strings.Contains(r.Problem, "could not be loaded") || !strings.Contains(r.Problem, "team.recipe.toml") {
		t.Errorf("an unreadable recipe is reported, not failed: %+v", r)
	}
}

func TestCheckProjectRefusesAnotherRecipe(t *testing.T) {
	s, p, rec := sessionFor(t, strings.Replace(recipeAt("1.2.0", "", "ruff", 100), `name = "team"`, `name = "other"`, 1))
	r := check(t, s, p, rec)
	if r.Drifted() || !strings.Contains(r.Problem, `the recipe "other"`) {
		t.Errorf("a different recipe cannot say how the project differs: %+v", r)
	}
}

func TestCheckProjectNeedsARecordedStack(t *testing.T) {
	s, p, rec := sessionFor(t, recipeAt("1.2.0", "", "ruff", 100))
	p.Stack = ""
	if r := check(t, s, p, rec); !strings.Contains(r.Problem, "records no stack") {
		t.Errorf("%+v", r)
	}
	p.Stack = "gone"
	if r := check(t, s, p, rec); !strings.Contains(r.Problem, "gone") {
		t.Errorf("%+v", r)
	}
}
