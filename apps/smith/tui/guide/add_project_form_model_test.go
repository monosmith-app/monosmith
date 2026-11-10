package guide

import (
	"strings"
	"testing"
)

func TestTypingNarrowsTheLanguages(t *testing.T) {
	m := keys(menu(), "down", "enter", "enter", "g", "o")
	view := m.View()
	if !strings.Contains(view, "Go") || strings.Contains(view, "TypeScript") {
		t.Fatalf("only Go should be left:\n%s", view)
	}
	m = keys(m, "enter", "s", "d", "k", "enter")
	if got := strings.Join(m.(Model).Argv, " "); got != "add app sdk --lang go" {
		t.Fatalf("argv %q", got)
	}
	m = keys(menu(), "down", "enter", "enter", "z", "z")
	if !strings.Contains(m.View(), "nothing matches") {
		t.Fatal("no match is said")
	}
	m = keys(m, "enter")
	if m.(Model).Done() {
		t.Fatal("enter with no match must not choose")
	}
	m = keys(m, "esc")
	if m.(Model).Title() == "" || !strings.Contains(m.View(), "TypeScript") {
		t.Fatal("esc clears the filter and stays on the question")
	}
}

func TestLanguageStartsOnTheWorkspaceMainLanguage(t *testing.T) {
	m := menu().WithDefault("lang", "go")
	m = keys(m, "down", "enter", "enter", "enter", "s", "d", "k", "enter").(Model)
	if got := strings.Join(m.Argv, " "); got != "add app sdk --lang go" {
		t.Fatalf("argv %q", got)
	}
}

func TestBadNameIsSaidWhileTypingAndRefused(t *testing.T) {
	m := keys(menu(), "down", "enter", "down", "down", "enter", "enter", "B", "a", "d")
	if view := m.View(); !strings.Contains(view, "lowercase words joined by dashes") || !strings.Contains(view, "my-package") {
		t.Fatalf("the problem shows live:\n%s", view)
	}
	m = keys(m, "enter")
	if m.(Model).Done() || len(m.(Model).Argv) > 0 {
		t.Fatal("a bad name must not go on")
	}
	m = keys(menu(), "down", "enter", "down", "down", "enter", "enter", "o", "k")
	if !strings.Contains(m.View(), "that name works") {
		t.Fatal("a good name is confirmed")
	}
}

func TestMatchesFilter(t *testing.T) {
	o := Option{Label: "C# / .NET", Value: "dotnet", Hint: "C# · .NET apps"}
	for typed, want := range map[string]bool{"dot": true, "C#": true, "NET apps": true, "rust": false, "": true} {
		if MatchesFilter(o, typed) != want {
			t.Errorf("%q: want %v", typed, want)
		}
	}
}

func TestFilteringByAFrameworkFindsItsLanguage(t *testing.T) {
	node := Option{Label: "Node.js / TypeScript", Value: "node", Hint: "web apps"}
	py := Option{Label: "Python", Value: "python", Hint: "scripts"}
	offers := []ChoiceOffer{
		{Axis: "framework", Name: "react", Language: "node"},
		{Axis: "framework", Name: "fastapi", Language: "python"},
	}
	base := Action{Steps: []Step{{Key: "lang", Options: []Option{node, py}, Filter: true}}, Argv: func(map[string][]string) []string { return nil }}
	opts := AddProjectAction(base, offers, nil).Steps[0].Options
	for typed, want := range map[string]string{"react": "node", "fastapi": "python", "fast": "python"} {
		var got []string
		for _, o := range opts {
			if MatchesFilter(o, typed) {
				got = append(got, o.Value)
			}
		}
		if len(got) != 1 || got[0] != want {
			t.Errorf("%q finds %v, want %s", typed, got, want)
		}
	}
	if via := MatchedVia(opts[0], "react"); len(via) != 1 || via[0] != "react" {
		t.Errorf("the match must say it came from react: %v", via)
	}
	if via := MatchedVia(opts[0], "node"); via != nil {
		t.Errorf("a plain match needs no note: %v", via)
	}
}

func TestNothingMatchesSuggestsTheClosestNames(t *testing.T) {
	opts := []Option{{Label: "Python", Value: "python", Also: []string{"fastapi", "flask"}}, {Label: "Node", Value: "node", Also: []string{"react", "vue"}}}
	if got := NothingMatches(opts, "reakt"); !strings.Contains(got, "Did you mean react?") {
		t.Errorf("%q", got)
	}
	if got := NothingMatches(opts, "zzzzzz"); !strings.Contains(got, "such as fastapi, flask, react") {
		t.Errorf("%q", got)
	}
}
