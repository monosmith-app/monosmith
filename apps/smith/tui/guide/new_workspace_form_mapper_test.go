package guide

import (
	"reflect"
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
	want := []string{"name", "ci", "language", "architecture", "more"}
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
