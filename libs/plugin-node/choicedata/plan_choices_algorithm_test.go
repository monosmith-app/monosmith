package choicedata

import (
	"reflect"
	"strings"
	"testing"
)

func TestPlanMergesTheBaseAndTheChosenChoicesInOrder(t *testing.T) {
	s, err := Load(fixture(nil), "d")
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Plan([]string{"react", "eslint", "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Applied, []string{"react", "bootstrap"}) {
		t.Errorf("applied %v: eslint is another plugin's", p.Applied)
	}
	if p.Dependencies["react"] != "^19.0.0" || p.Dependencies["bootstrap"] != "^5.3.0" || p.DevDependencies["vite"] != "^8.0.0" {
		t.Errorf("versions come from the catalog: %+v %+v", p.Dependencies, p.DevDependencies)
	}
	if p.Tasks["build"] != "vite build" {
		t.Errorf("tasks %v", p.Tasks)
	}
	if !reflect.DeepEqual(p.Slots["vite.plugins"], []string{"react()"}) || len(p.Slots["entry.imports"]) != 1 {
		t.Errorf("slots %v", p.Slots)
	}
	if !strings.Contains(p.Files["src/main.ts"], "react") {
		t.Errorf("a choice replaces the base's file: %q", p.Files["src/main.ts"])
	}
}

func TestPlanOfNoChoicesIsTheBase(t *testing.T) {
	s, _ := Load(fixture(nil), "d")
	p, err := s.Plan(nil)
	if err != nil || len(p.Applied) != 0 || p.Tasks["build"] == "" || !strings.Contains(p.Files["src/main.ts"], "base") {
		t.Fatalf("%+v %v", p, err)
	}
}

func TestPlanRefusesTwoChoicesThatWriteOneFileUnlessOneRequiresTheOther(t *testing.T) {
	other := "axis = \"style\"\nname = \"other\"\nlanguage = \"node\"\n[files]\n\"src/main.ts\" = \"x\"\n"
	s, err := Load(fixture(map[string]string{"other.toml": other}), "d")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Plan([]string{"react", "other"})
	if err == nil || !strings.Contains(err.Error(), "react and other both set the file src/main.ts, and other does not require react") {
		t.Fatalf("got %v", err)
	}
	// requires is a top-level member: it goes before the [files] table.
	requiring := "axis = \"style\"\nname = \"other\"\nlanguage = \"node\"\nrequires = [\"react\"]\n[files]\n\"src/main.ts\" = \"x\"\n"
	s, err = Load(fixture(map[string]string{"other.toml": requiring}), "d")
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Plan([]string{"react", "other"})
	if err != nil || p.Files["src/main.ts"] != "x" {
		t.Fatalf("a choice that requires the other may replace its file: %v %v", p.Files, err)
	}
}
