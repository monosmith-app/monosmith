package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

func TestTheAddPlanSaysTheChoicesInWordsAndMarksTheDefaults(t *testing.T) {
	offers, err := guide.ParseChoices(`[
 {"axis":"framework","name":"react","title":"React","description":"a React app","language":"node"},
 {"axis":"lint","name":"eslint","title":"ESLint","description":"lint","language":"node","default":true},
 {"axis":"shell","name":"tauri","title":"Tauri","description":"a shell","language":"node"},
 {"axis":"style","name":"bootstrap","title":"Bootstrap","description":"styles","language":"node"},
 {"axis":"test","name":"vitest","title":"Vitest","description":"tests","language":"node","default":true},
 {"axis":"test","name":"jest","title":"Jest","description":"tests","language":"node"}]`)
	if err != nil {
		t.Fatal(err)
	}
	got := addPlanWhy([]string{"add", "app", "web", "--lang", "node", "--with", "react,eslint,vitest"}, 1, offers)
	want := "It is built from your choices: React, ESLint (default), Vitest (default); no desktop shell, no style."
	if !strings.Contains(got, want) {
		t.Fatalf("want %q in %q", want, got)
	}
}

func TestTheAddPlanNamesTheChoicesAndTheSiblings(t *testing.T) {
	for _, tc := range []struct {
		argv     []string
		projects int
		want     []string
		not      []string
	}{
		{[]string{"add", "app", "web", "--with", "react,vitest"}, 4, []string{"your choices: react, vitest", "follows the other projects"}, []string{"defaults"}},
		{[]string{"add", "app", "web", "--fresh"}, 4, []string{"stack's own defaults", "does not follow the other projects"}, []string{"follows the other"}},
		{[]string{"add", "app", "web"}, 1, []string{"stack's own defaults"}, []string{"follow"}},
	} {
		got := addPlanWhy(tc.argv, tc.projects, nil)
		for _, w := range tc.want {
			if !strings.Contains(got, w) {
				t.Errorf("%v: missing %q in %q", tc.argv, w, got)
			}
		}
		for _, n := range tc.not {
			if strings.Contains(got, n) {
				t.Errorf("%v: unexpected %q in %q", tc.argv, n, got)
			}
		}
	}
}
