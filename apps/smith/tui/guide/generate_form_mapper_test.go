package guide

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
)

func described() clicontract.CommandDescription {
	return clicontract.CommandDescription{
		Name: "uninstall", Description: "Remove a package from projects",
		Arguments: []clicontract.ArgumentDescription{{Name: "template", Required: false}, {Name: "packages", Required: true, Variadic: true}},
		Options: []clicontract.OptionDescription{
			{Name: "json", TakesValue: "none"}, {Name: "force", TakesValue: "none", Description: "overwrite"},
			{Name: "workspace", TakesValue: "required", Repeatable: true, Description: "the workspaces"},
			{Name: "ci", TakesValue: "required"},
		},
	}
}

func TestFormForAsksRequiredFirstThenSwitchesThenValuedOptions(t *testing.T) {
	a := FormFor(described(), []string{"uninstall"}, "uninstall")
	var keys []string
	for _, s := range a.Steps {
		keys = append(keys, s.Key)
	}
	want := []string{"arg:packages", "arg:template", "switches", "flag:workspace", "flag:ci"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("steps %v, want %v", keys, want)
	}
	if a.Steps[0].Optional || !a.Steps[1].Optional || !a.Steps[3].Optional {
		t.Errorf("optional flags wrong: %+v", a.Steps)
	}
	if got := a.Steps[2].Options; len(got) != 2 || got[0].Label != "none" || got[1].Value != "--force" {
		t.Errorf("switches %+v: want none, then --force (json is never offered)", got)
	}
	if a.Presents != "uninstall" || a.Title != "Remove a package from projects" || !a.NeedsWorkspace {
		t.Errorf("action %+v", a)
	}
}

func TestFormForBuildsTheCommandLineFromTheAnswers(t *testing.T) {
	a := FormFor(described(), []string{"uninstall"}, "uninstall")
	cases := []struct {
		answers map[string][]string
		want    []string
	}{
		{map[string][]string{"arg:packages": {"zod  lodash"}, "arg:template": {""}, "switches": {""}, "flag:workspace": {""}, "flag:ci": {""}},
			[]string{"uninstall", "zod", "lodash"}},
		{map[string][]string{"arg:packages": {"zod"}, "arg:template": {"e2e"}, "switches": {"--force"}, "flag:workspace": {"web api"}, "flag:ci": {"github"}},
			[]string{"uninstall", "zod", "e2e", "--force", "--workspace", "web", "--workspace", "api", "--ci", "github"}},
	}
	for _, c := range cases {
		if got := a.Argv(c.answers); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v = %v, want %v", c.answers, got, c.want)
		}
	}
}

func TestAnOptionalStepMayBeLeftEmpty(t *testing.T) {
	a := FormFor(clicontract.CommandDescription{Name: "pipeline", Description: "Add a pipeline", Arguments: []clicontract.ArgumentDescription{{Name: "template"}}}, []string{"pipeline"}, "pipeline")
	m, ok := NewModel([]Action{a}).Pick("Add a pipeline")
	if !ok {
		t.Fatal("the form is not in the menu")
	}
	m = keys(m, "enter").(Model)
	if !m.Done() || !reflect.DeepEqual(m.Argv, []string{"pipeline"}) {
		t.Errorf("done %v, argv %v", m.Done(), m.Argv)
	}
}

func TestFormForASubcommandRunsItUnderItsParent(t *testing.T) {
	up := clicontract.CommandDescription{Name: "up", Description: "Update dependencies", Arguments: []clicontract.ArgumentDescription{{Name: "packages", Required: true, Variadic: true}}}
	a := FormFor(up, []string{"deps", "up"}, "deps up")
	if a.Hint != "smith deps up" || a.Presents != "deps up" {
		t.Errorf("action %+v", a)
	}
	if got := a.Argv(map[string][]string{"arg:packages": {"zod lodash"}}); !reflect.DeepEqual(got, []string{"deps", "up", "zod", "lodash"}) {
		t.Errorf("argv %v", got)
	}
}
