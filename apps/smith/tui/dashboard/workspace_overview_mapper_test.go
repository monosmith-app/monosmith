package dashboard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTheProjectsRowNamesTheMostUsedLanguagesInWords(t *testing.T) {
	langs := []string{"node", "node", "node", "python", "jvm", "jvm", "go", "rust", "dotnet"}
	got := projectsText(Facts{Projects: len(langs), Languages: langs})
	want := "9 projects: 3 TypeScript / JavaScript · 2 Java / Kotlin · 1 Python · 1 Go · 2 more"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := projectsText(Facts{Projects: 1}); got != "1 project" {
		t.Fatalf("no languages: %q", got)
	}
}

func TestHomeFillsItsSpaceWithProjectsAndWhereToGo(t *testing.T) {
	f := Facts{ProjectsLoaded: true, Checked: true, Projects: 2, Languages: []string{"go", "go"}}
	lines, _, _ := Lines(theme.Default(), f, 100)
	var plain []string
	for _, l := range lines {
		plain = append(plain, pane.Strip(l))
	}
	all := strings.Join(plain, "\n")
	for _, want := range []string{"2 projects: 2 Go", "Where to go", "Doctor", "Release"} {
		if !strings.Contains(all, want) {
			t.Fatalf("Home lacks %q:\n%s", want, all)
		}
	}
}
