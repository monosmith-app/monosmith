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
