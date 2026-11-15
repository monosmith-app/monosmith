package dashboard

import (
	"strings"
	"testing"
)

func TestTheWorkspaceLineNamesTheMostUsedLanguagesInWords(t *testing.T) {
	langs := []string{"node", "node", "node", "python", "jvm", "jvm", "go", "rust", "dotnet"}
	got := projectsText(Facts{Projects: len(langs), Languages: langs})
	want := "9 projects (3 TypeScript / JavaScript, 2 Java / Kotlin, 1 Python, 3 more)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := projectsText(Facts{Projects: 1}); got != "1 project" {
		t.Fatalf("no languages: %q", got)
	}
}

func TestTheSummaryLineIsTheFirstLineAndHasNoNameOrBranch(t *testing.T) {
	f := Facts{ProjectsLoaded: true, Projects: 2, Languages: []string{"go", "go"}, Checked: true, DepsChecked: true}
	text := plainLines(Page(plainTheme(), f, 100, Away))
	all := strings.Join(text, "\n")
	if len(text) < 2 || strings.TrimSpace(text[1]) != "2 projects (2 Go)" {
		t.Fatalf("the first line is not the summary:\n%s", all)
	}
	if strings.Count(all, "projects (") != 1 {
		t.Fatalf("the summary shows more than once:\n%s", all)
	}
}
