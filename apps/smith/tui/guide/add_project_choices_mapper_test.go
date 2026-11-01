package guide

import (
	"reflect"
	"strings"
	"testing"
)

const offersJSON = `[
 {"axis":"framework","name":"react","description":"a React app","language":"node"},
 {"axis":"framework","name":"svelte","description":"a Svelte app","language":"node","problem":"the node plugin is off"},
 {"axis":"test","name":"vitest","description":"unit tests","language":"node","default":true},
 {"axis":"style","name":"bootstrap","description":"styles","language":"node","requires":["react"]},
 {"axis":"test","name":"gotest","description":"go tests","language":"go"}
]`

func TestParseChoicesKeepsTheUsableOnesAndRefusesOtherText(t *testing.T) {
	offers, err := ParseChoices(offersJSON)
	if err != nil || len(offers) != 4 {
		t.Fatalf("%v %+v", err, offers)
	}
	for _, o := range offers {
		if o.Name == "svelte" {
			t.Fatal("a choice with a problem cannot be offered")
		}
	}
	if _, err := ParseChoices("✖ not a workspace"); err == nil || !strings.Contains(err.Error(), "did not print a list") {
		t.Fatalf("plain text must be refused with a reason: %v", err)
	}
}

// press sends key names to a form and returns it.
func press(m Model, ks ...string) Model { return keys(m, ks...).(Model) }

func plainAdd() Action {
	for _, a := range Actions([]string{"go", "node"}, nil, nil) {
		if a.Title == "Add a project" {
			return a
		}
	}
	panic("no Add a project")
}

func TestAddProjectAsksTheChoicesOfTheLanguagePickedOnly(t *testing.T) {
	offers, _ := ParseChoices(offersJSON)
	a := AddProjectAction(plainAdd(), offers, nil)
	m := NewModel([]Action{a})
	m, ok := m.Pick("Add a project")
	if !ok {
		t.Fatal("no form")
	}
	m = press(m, "enter")                     // an app
	m = press(m, "n", "o", "d", "e", "enter") // node
	m = press(m, "w", "e", "b", "enter")
	view := m.View()
	for _, want := range []string{"Build it from choices", "framework: react", "styles (needs react)", "test: vitest"} {
		if !strings.Contains(view, want) {
			t.Fatalf("the node choices lack %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "gotest") || strings.Contains(view, "svelte") {
		t.Fatalf("another language's or an unusable choice is offered:\n%s", view)
	}
	m = press(m, "space", "down", "down", "down", "space", "enter")
	if got := strings.Join(m.Argv, " "); got != "add app web --lang node --with react,vitest" {
		t.Fatalf("argv %q", got)
	}
}

func TestAddProjectSkipsTheChoicesWhenNothingIsTicked(t *testing.T) {
	offers, _ := ParseChoices(offersJSON)
	m, _ := NewModel([]Action{AddProjectAction(plainAdd(), offers, nil)}).Pick("Add a project")
	m = press(m, "enter", "n", "o", "d", "e", "enter", "w", "e", "b", "enter", "enter")
	if !m.Done() || strings.Join(m.Argv, " ") != "add app web --lang node" {
		t.Fatalf("done=%v argv %q", m.Done(), m.Argv)
	}
}

func TestAddProjectOffersFreshOnlyWhereThereAreSiblingsToFollow(t *testing.T) {
	plain := plainAdd()
	if got := AddProjectAction(plain, nil, []string{"a", "b"}); len(got.Steps) != len(plain.Steps) {
		t.Fatal("two projects follow nobody: no extra question")
	}
	a := AddProjectAction(plain, nil, []string{"a", "b", "c"})
	if len(a.Steps) != len(plain.Steps)+1 || a.Steps[len(a.Steps)-1].Key != "follow" {
		t.Fatalf("three projects make a question about following them: %+v", a.Steps)
	}
	if got := a.Argv(map[string][]string{"kind": {"lib"}, "name": {"core"}, "lang": {"go"}, "follow": {"fresh"}}); !reflect.DeepEqual(got, []string{"add", "lib", "core", "--lang", "go", "--fresh"}) {
		t.Fatalf("argv %v", got)
	}
	if got := a.Argv(map[string][]string{"kind": {"lib"}, "name": {"core"}, "lang": {"go"}, "follow": {"follow"}}); len(got) != 5 {
		t.Fatalf("following the siblings is the default and adds no flag: %v", got)
	}
}

func TestWithActionReplacesTheActionOfTheSameTitle(t *testing.T) {
	m := NewModel([]Action{{Title: "A"}, {Title: "B"}}).WithAction(Action{Title: "A", Hint: "new"}).WithAction(Action{Title: "C"})
	if a, _ := m.ActionTitled("A"); a.Hint != "new" {
		t.Fatalf("A was not replaced: %+v", a)
	}
	if _, ok := m.ActionTitled("C"); !ok {
		t.Fatal("C was not added")
	}
	if _, ok := m.ActionTitled("D"); ok {
		t.Fatal("D does not exist")
	}
}

func TestPluginSearchBecomesAFormThatInstallsThePluginPicked(t *testing.T) {
	hits, err := ParsePluginHits(`[{"name":"deployer","newest":"0.2.0","description":"Deploys to the moon","roles":["deploy"]},{"name":"widget","newest":"1.0.0","description":"Scaffolds"}]`)
	if err != nil || len(hits) != 2 {
		t.Fatalf("%v %+v", err, hits)
	}
	a := PluginInstallAction(hits)
	m, _ := NewModel([]Action{a}).Pick(a.Title)
	if view := m.View(); !strings.Contains(view, "v0.2.0 · Deploys to the moon [deploy]") {
		t.Fatalf("the hint says the version, the purpose and the roles:\n%s", view)
	}
	m = press(m, "down", "enter")
	if got := strings.Join(m.Argv, " "); got != "plugin install widget" {
		t.Fatalf("argv %q", got)
	}
	if _, err := ParsePluginHits("No plugin matches"); err == nil {
		t.Fatal("text is not a list")
	}
}

func TestWhyAsksForAPath(t *testing.T) {
	var why Action
	for _, a := range Actions(nil, nil, nil) {
		if a.Presents == "why" {
			why = a
		}
	}
	if why.Title == "" || why.Argv(map[string][]string{"path": {"apps/web/tsconfig.json"}})[1] != "apps/web/tsconfig.json" {
		t.Fatalf("no form for smith why: %+v", why)
	}
}
