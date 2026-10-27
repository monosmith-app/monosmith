package architecture

import (
	"bytes"
	"strings"
	"testing"
)

var members = []Member{{"api", "go"}, {"web", "node"}, {"infra", "bicep"}}

func TestEveryArchitectureSaysWhichLanguagesItAppliesTo(t *testing.T) {
	for _, a := range Catalog() {
		if a.Table == "" {
			if len(a.AppliesTo()) != 0 || a.AppliesText() != "" {
				t.Fatalf("%s applies to nothing, it is the absence of a choice", a.Name)
			}
			continue
		}
		if len(a.AppliesTo()) == 0 || !strings.Contains(a.AppliesText(), "go") {
			t.Errorf("%s declares %v", a.Name, a.AppliesTo())
		}
	}
}

func TestAnArchitectureChecksTheProjectsOfItsLanguagesAndSkipsTheRest(t *testing.T) {
	a, err := Find("hexagonal")
	if err != nil {
		t.Fatal(err)
	}
	c := a.Cover(members)
	if len(c.Checked) != 2 || len(c.Skipped) != 1 || c.Skipped[0].Name != "infra" {
		t.Fatalf("%+v", c)
	}
	if got := c.Here(); got != "checks api (go), web (node); skips infra (bicep) (its language is not covered)" {
		t.Fatalf("%q", got)
	}
	if a.Warning(c) != "" {
		t.Fatal("it checks something: no warning")
	}
}

func TestAnArchitectureThatCoversNoProjectWarns(t *testing.T) {
	a, _ := Find("hexagonal")
	c := a.Cover([]Member{{"infra", "bicep"}})
	warning := a.Warning(c)
	for _, want := range []string{"hexagonal applies to", "infra (bicep)", "would check nothing"} {
		if !strings.Contains(warning, want) {
			t.Fatalf("%q lacks %q", warning, want)
		}
	}
}

func TestAWorkspaceWithoutProjectsOrWithNoneHasNothingToWarnAbout(t *testing.T) {
	a, _ := Find("hexagonal")
	if a.Warning(a.Cover(nil)) != "" {
		t.Fatal("an empty workspace is not a gap")
	}
	none, _ := Find(None)
	if c := none.Cover(members); len(c.Checked)+len(c.Skipped) != 0 || none.Warning(c) != "" || c.Here() != "" {
		t.Fatalf("none checks nothing on purpose: %+v", c)
	}
}

func TestLongListsOfProjectsAreCounted(t *testing.T) {
	var many []Member
	for _, n := range []string{"a", "b", "c", "d", "e", "f"} {
		many = append(many, Member{n, "bicep"})
	}
	a, _ := Find("clean")
	if w := a.Warning(a.Cover(many)); !strings.Contains(w, "and 2 more") {
		t.Fatal(w)
	}
}

func TestTheListShowsWhatEachChoiceAppliesToAndDoesHere(t *testing.T) {
	var out bytes.Buffer
	current, _ := Find("hexagonal")
	List(&out, current, members)
	got := out.String()
	for _, want := range []string{"* hexagonal", "  clean", "applies to: ", "here: checks api (go), web (node); skips infra (bicep)"} {
		if !strings.Contains(got, want) {
			t.Errorf("the list lacks %q:\n%s", want, got)
		}
	}
	out.Reset()
	List(&out, current, nil)
	if strings.Contains(out.String(), "here:") {
		t.Fatalf("no projects known, nothing to say about them:\n%s", out.String())
	}
}
