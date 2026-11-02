package guide

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

func TestTheFormAsksWhatSmithNewAsks(t *testing.T) {
	var formKeys, commandKeys []string
	for _, s := range NewWorkspaceSteps() {
		formKeys = append(formKeys, s.Key)
	}
	commandKeys = append(commandKeys, "name")
	for _, q := range scaffold.NewQuestions() {
		commandKeys = append(commandKeys, q.Key)
	}
	if !reflect.DeepEqual(formKeys, commandKeys) {
		t.Fatalf("form %v, command %v", formKeys, commandKeys)
	}
}

func TestDefaultsBuildTheShortestCommand(t *testing.T) {
	got := NewWorkspaceArgv(map[string][]string{"name": {"acme"}, "ci": {"github"}, "language": {"later"}, "architecture": {"vertical-slices"}, "more": {"no"}})
	if want := []string{"new", "acme", "--yes"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestAnswersBecomeFlags(t *testing.T) {
	got := NewWorkspaceArgv(map[string][]string{
		"name": {"acme"}, "ci": {"azure"}, "language": {"node"}, "node-stack": {"node-vitest"},
		"architecture": {"none"}, "more": {"yes"}, "release-branch": {"trunk"}, "release-remote": {""},
		"github-releases": {"yes"}, "cache-dir": {""}, "cache-max-mb": {"0"},
	})
	want := []string{"new", "acme", "--yes", "--ci", "azure", "--language", "node", "--default-stack", "node=node-vitest",
		"--architecture", "none", "--release-branch", "trunk", "--github-releases", "--cache-max-mb", "0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestQuestionsThatDoNotApplyAreSkipped(t *testing.T) {
	var action Action
	for _, a := range Actions(nil, nil, nil) {
		if a.Presents == "new" {
			action = a
		}
	}
	m := NewModel([]Action{action})
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyEnter}) // choose the action: asks the name
	m.input.SetValue("acme")
	enter := tea.KeyMsg{Type: tea.KeyEnter}
	var seen []string
	for i := 0; i < 20 && !m.Done(); i++ {
		seen = append(seen, m.current().Key)
		m, _ = m.Handle(enter)
	}
	want := []string{"name", "ci", "language", "architecture", "agent", "devcontainer", "more"}
	if !reflect.DeepEqual(seen, want) || !m.Done() {
		t.Fatalf("with every default the form asks %v, got %v (done %v)", want, seen, m.Done())
	}
	if !reflect.DeepEqual(m.Argv, []string{"new", "acme", "--yes"}) {
		t.Errorf("%v", m.Argv)
	}
}

func TestSavingARecipeFromTheFormUsesTheWorkspaceNameByDefault(t *testing.T) {
	got := NewWorkspaceArgv(map[string][]string{"name": {"acme"}, "language": {"go"}, "save-recipe": {"yes"}, "recipe-name": {""}})
	if want := []string{"new", "acme", "--yes", "--language", "go", "--save-recipe", "acme"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
	got = NewWorkspaceArgv(map[string][]string{"name": {"acme"}, "language": {"go"}, "save-recipe": {"yes"}, "recipe-name": {"team"}})
	if got[len(got)-1] != "team" {
		t.Fatalf("%v", got)
	}
	if got := NewWorkspaceArgv(map[string][]string{"name": {"acme"}, "save-recipe": {"no"}}); len(got) != 3 {
		t.Fatalf("no means no flag: %v", got)
	}
}

func newWorkspaceAction(t *testing.T) Action {
	t.Helper()
	for _, a := range Actions(nil, nil, nil) {
		if a.Presents == "new" {
			return a
		}
	}
	t.Fatal("no Create a workspace action")
	return Action{}
}

func TestAWorkspaceNameThatSmithRefusesIsRefusedOnTheSpot(t *testing.T) {
	m, _ := NewModel([]Action{newWorkspaceAction(t)}).Handle(tea.KeyMsg{Type: tea.KeyEnter})
	for _, bad := range []string{"my shop", "My-Shop", "-shop", "shop_1"} {
		m.input.SetValue(bad)
		m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyEnter})
		if m.current().Key != "name" || !strings.Contains(m.err, "lowercase words joined by dashes") {
			t.Fatalf("%q: step %q, error %q", bad, m.current().Key, m.err)
		}
	}
	m.input.SetValue("my-shop")
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyEnter})
	if m.current().Key == "name" {
		t.Fatalf("a good name must move on")
	}
}

func TestResumeKeepsTheAnswersAndShowsTheProblem(t *testing.T) {
	answers := map[string][]string{"name": {"acme"}, "ci": {"azure"}, "language": {"go"}}
	m, ok := NewModel([]Action{newWorkspaceAction(t)}).Resume("Create a workspace", answers, "acme already holds a smith workspace")
	if !ok || m.err != "acme already holds a smith workspace" || m.input.Value() != "acme" {
		t.Fatalf("ok %v, error %q, name %q", ok, m.err, m.input.Value())
	}
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyEnter})
	if step := m.current(); step.Key != "ci" || step.Options[m.cursor].Value != "azure" {
		t.Fatalf("the ci question must start on the earlier answer: cursor %d", m.cursor)
	}
}

func TestARefusedNameOffersTheCorrectedOneAndTabTakesIt(t *testing.T) {
	m, _ := NewModel([]Action{newWorkspaceAction(t)}).Handle(tea.KeyMsg{Type: tea.KeyEnter})
	m.input.SetValue("My Shop")
	if view := m.View(); !strings.Contains(view, "tab    use my-shop") || !strings.Contains(view, "folder, a package name") {
		t.Fatalf("the choices and the rule must show: %s", view)
	}
	m, _ = m.Handle(tea.KeyMsg{Type: tea.KeyTab})
	if m.input.Value() != "my-shop" {
		t.Fatalf("tab must take the suggestion: %q", m.input.Value())
	}
	if got := suggestName("workspace")(nil, "9 lives!"); got != "lives" {
		t.Errorf("%q", got)
	}
	if got := suggestName("workspace")(nil, "###"); got != "" {
		t.Errorf("%q", got)
	}
}
