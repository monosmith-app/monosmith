package recipewizard

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspacerecipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// workspaceFacts are a workspace of two Node projects on a jest stack and one
// Python project, as workspacerecipe.Load would read them.
func workspaceFacts() *workspacerecipe.Facts {
	nodeJest := protocol.Stack{Language: "node", Toolchain: []string{"eslint", "jest"}, Structure: []string{"slices"}, Deploy: []string{"docker"}}
	return &workspacerecipe.Facts{
		Projects: []workspacerecipe.ProjectFact{
			{Language: "node", StackName: "node-jest", Stack: nodeJest},
			{Language: "node", StackName: "node-jest", Stack: nodeJest},
			{Language: "python"},
		},
		Provides: map[string][]string{"eslint": {"lint"}, "jest": {"test"}},
	}
}

// fromWorkspace runs `smith recipe new --from-workspace` in a new workspace
// and returns its folder, the output and the error.
func fromWorkspace(t *testing.T, o Options) (string, string, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("[workspace]\nname = \"lab\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	// smith ci names the workspace it runs in; the test's workspace is dir.
	t.Setenv("SMITH_WORKSPACE_ROOT", dir)
	var out bytes.Buffer
	o.FromWorkspace, o.Facts, o.Lookup, o.Out = true, workspaceFacts(), fakeLookup, &out
	if o.In == nil {
		o.In = strings.NewReader("")
	}
	err := Run(o)
	return dir, out.String(), err
}

func TestFromWorkspaceWritesARecipeThatFixesWhatTheWorkspaceUses(t *testing.T) {
	dir, out, err := fromWorkspace(t, Options{Name: "team-node"})
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	written, err := os.ReadFile(filepath.Join(dir, "recipes", "team-node.recipe.toml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`language = "node"`, `lint = "eslint"`, `test = "jest"`, `version = "0.1.0"`} {
		if !strings.Contains(string(written), want) {
			t.Errorf("the recipe lacks %s:\n%s", want, written)
		}
	}
	if !strings.Contains(out, "the stack node-jest") || !strings.Contains(out, "deploy") {
		t.Errorf("the output does not show what is fixed:\n%s", out)
	}
	doc, err := recipe.ParseRecipe(written, recipe.Source{Ref: "team-node.recipe.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recipe.ValidateAll(recipe.Loaded{Chain: []recipe.Document{doc}}, fakeLookup); err != nil {
		t.Errorf("the recipe does not check out: %v", err)
	}
}

func TestFromWorkspaceRefusesToReplaceARecipeWithoutForce(t *testing.T) {
	dir, _, err := fromWorkspace(t, Options{Name: "team-node"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	o := Options{FromWorkspace: true, Facts: workspaceFacts(), Lookup: fakeLookup, Out: &out, In: strings.NewReader("y\n"), Name: "team-node", Language: "node", Interactive: true}
	if err := Run(o); !errors.Is(err, ErrExists) || !strings.Contains(err.Error(), "--force") {
		t.Errorf("err = %v; an existing recipe must be refused without --force", err)
	}
	o.Force = true
	if err := Run(o); err != nil {
		t.Errorf("--force: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "recipes", "team-node.recipe.toml")); err != nil {
		t.Error(err)
	}
}

func TestFromWorkspaceAsksTheLanguageTheNameAndTheVersionOnATerminal(t *testing.T) {
	in := strings.NewReader("python\nBad Name\nteam-py\n1.2.0\n")
	dir, out, err := fromWorkspace(t, Options{Interactive: true, In: in})
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	written, err := os.ReadFile(filepath.Join(dir, "recipes", "team-py.recipe.toml"))
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(string(written), `language = "python"`) || !strings.Contains(string(written), `version = "1.2.0"`) {
		t.Errorf("recipe:\n%s", written)
	}
	if !strings.Contains(out, "node, python") || !strings.Contains(out, "lowercase words") {
		t.Errorf("the questions do not show the options or the problem:\n%s", out)
	}
}

func TestFromWorkspaceNeedsANameWithoutATerminalAndGoesAlone(t *testing.T) {
	if _, _, err := fromWorkspace(t, Options{}); err == nil || !strings.Contains(err.Error(), "--name") {
		t.Errorf("no name: %v", err)
	}
	if _, _, err := fromWorkspace(t, Options{Name: "team", Fixed: []string{"language=node"}}); err == nil {
		t.Error("--fixed was taken next to --from-workspace")
	}
	if err := Run(Options{Language: "go", Name: "x", Out: &bytes.Buffer{}, In: strings.NewReader("")}); err == nil {
		t.Error("--language was taken without --from-workspace")
	}
}
