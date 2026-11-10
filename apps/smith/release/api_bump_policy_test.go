package release

import (
	"slices"
	"strings"
	"testing"
)

func TestDiffAPIListsRemovedAndAddedNames(t *testing.T) {
	got := DiffAPI([]string{"a", "b", "c"}, []string{"b", "c", "d", "e"})
	if !slices.Equal(got.Removed, []string{"a"}) || !slices.Equal(got.Added, []string{"d", "e"}) {
		t.Errorf("DiffAPI = %+v", got)
	}
	if same := DiffAPI([]string{"a"}, []string{"a"}); same.Needs(Version{Major: 1}) != None {
		t.Errorf("no difference must ask for nothing, got %+v", same)
	}
}

func TestCheckAPIBumpClassifiesTheChange(t *testing.T) {
	cases := []struct {
		name    string
		current string
		planned string
		change  APIChange
		needs   Level
		propose string
	}{
		{"a fix that removes an export needs a major", "1.2.3", "patch", APIChange{Removed: []string{"x"}}, Major, "2.0.0"},
		{"a feature that removes an export needs a major", "1.2.3", "minor", APIChange{Removed: []string{"x"}}, Major, "2.0.0"},
		{"a fix that adds an export needs a minor", "1.2.3", "patch", APIChange{Added: []string{"y"}}, Minor, "1.3.0"},
		{"before 1.0 a removal needs only a minor", "0.4.1", "patch", APIChange{Removed: []string{"x"}}, Minor, "0.5.0"},
		{"removal and addition together need the major", "2.0.0", "minor", APIChange{Removed: []string{"x"}, Added: []string{"y"}}, Major, "3.0.0"},
	}
	for _, c := range cases {
		f, ok := CheckAPIBump(Item{Project: "lib", Current: c.current, Level: c.planned, Next: "next"}, c.change)
		if !ok || f.Needs != c.needs || f.Proposed != c.propose {
			t.Errorf("%s: got %+v (found %v), want needs %s and %s", c.name, f, ok, c.needs, c.propose)
		}
	}
}

func TestCheckAPIBumpIsSilentWhenTheBumpIsEnough(t *testing.T) {
	cases := []struct {
		name   string
		item   Item
		change APIChange
	}{
		{"major already", Item{Current: "1.0.0", Level: "major"}, APIChange{Removed: []string{"x"}}},
		{"minor for an addition", Item{Current: "1.0.0", Level: "minor"}, APIChange{Added: []string{"y"}}},
		{"minor before 1.0 for a removal", Item{Current: "0.3.0", Level: "minor"}, APIChange{Removed: []string{"x"}}},
		{"nothing changed", Item{Current: "1.0.0", Level: "patch"}, APIChange{}},
		{"never released", Item{Current: "", Level: "patch"}, APIChange{Removed: []string{"x"}}},
		{"forced by a directive", Item{Current: "1.0.0", Level: "forced"}, APIChange{Removed: []string{"x"}}},
	}
	for _, c := range cases {
		if f, ok := CheckAPIBump(c.item, c.change); ok {
			t.Errorf("%s: want no finding, got %+v", c.name, f)
		}
	}
}

func TestAPIFindingSentenceNamesThePackageTheNamesAndTheBumpToUse(t *testing.T) {
	f, _ := CheckAPIBump(Item{Project: "ui", Current: "1.2.3", Level: "patch", Next: "1.2.4"},
		APIChange{Removed: []string{"a", "b", "c", "d", "e", "f", "g"}})
	want := "ui: the commits release a patch (1.2.4), but the public API lost 7 exports (a, b, c, d, e, and 2 more); " +
		"that needs a major (2.0.0). Mark a commit as breaking (`feat!:` or a `BREAKING CHANGE:` footer), or restore the export."
	if got := f.Sentence(); got != want {
		t.Errorf("sentence =\n%s\nwant\n%s", got, want)
	}
	add, _ := CheckAPIBump(Item{Project: "ui", Current: "1.2.3", Level: "patch", Next: "1.2.4"}, APIChange{Added: []string{"fresh"}})
	for _, part := range []string{"gained 1 export (fresh)", "needs a minor (1.3.0)", "`feat:`"} {
		if !strings.Contains(add.Sentence(), part) {
			t.Errorf("sentence %q lacks %q", add.Sentence(), part)
		}
	}
}
