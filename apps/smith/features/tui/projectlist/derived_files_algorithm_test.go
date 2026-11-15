package projectlist

import (
	"reflect"
	"strings"
	"testing"
)

func withDerived() Project {
	return Project{Name: "web", Derived: []DerivedFile{
		{Path: "apps/web/tsconfig.json", Status: StatusCurrent},
		{Path: "apps/web/jest.config.js", Status: StatusStale},
		{Path: "apps/web/package.json", Status: StatusDrift, Keys: true},
		{Path: "apps/web/.swcrc", Status: StatusDrift},
		{Path: "apps/web/vite.config.ts", Status: StatusDiverged},
		{Path: "apps/web/tsconfig.lib.json", Status: StatusDiffers},
		{Path: "apps/web/notes.md", Status: StatusEjected},
	}}
}

func TestDeriveArgvNamesTheFilesEachActionTakes(t *testing.T) {
	p := withDerived()
	for action, want := range map[string][]string{
		DeriveApply:   {"derive"},
		DeriveAdopt:   {"derive", "--adopt", "apps/web/package.json"},
		DeriveEject:   {"derive", "--eject", "apps/web/package.json", "--eject", "apps/web/.swcrc", "--eject", "apps/web/vite.config.ts"},
		DeriveReplace: {"derive", "--replace", "apps/web/package.json", "--replace", "apps/web/.swcrc", "--replace", "apps/web/vite.config.ts"},
		"unknown":     nil,
	} {
		if got := p.DeriveArgv(action); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, want %v", action, got, want)
		}
	}
	if got := p.DeriveArgv(DeriveExplain); len(got) != 1+2*len(p.Derived) || got[1] != "--explain" {
		t.Errorf("explain names every file: %v", got)
	}
}

func TestDeriveArgvIsNilWhenNoFileTakesTheAction(t *testing.T) {
	clean := Project{Derived: []DerivedFile{{Path: "a", Status: StatusCurrent}, {Path: "b", Status: StatusDrift}}}
	if clean.DeriveArgv(DeriveAdopt) != nil {
		t.Error("a drifted file that is not a keys file cannot be adopted")
	}
	if (Project{}).DeriveArgv(DeriveExplain) != nil {
		t.Error("a project with no derived file has nothing to explain")
	}
}

func TestOnlyTheFilesThatNeedActionAreCounted(t *testing.T) {
	if got := withDerived().DerivedNeedingAction(); got != 4 {
		t.Fatalf("stale, 2 drift and diverged need action, got %d", got)
	}
}

func TestEveryStatusHasAPlainSentence(t *testing.T) {
	for _, s := range []string{StatusCurrent, StatusStale, StatusMissing, StatusUnrecorded, StatusDrift, StatusDiverged, StatusBroken, StatusEjected, StatusDiffers} {
		w := StatusWords(s)
		if w == s || strings.TrimSpace(w) == "" {
			t.Errorf("%s has no plain words: %q", s, w)
		}
	}
	if StatusWords("new") != "new" {
		t.Error("an unknown status is shown as it is")
	}
}
