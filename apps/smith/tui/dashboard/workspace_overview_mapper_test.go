package dashboard

import "testing"

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

func TestTheWorkspaceLineStartsWithTheNameAndTheBranch(t *testing.T) {
	f := Facts{Workspace: "acme", Branch: "main", Projects: 2, Languages: []string{"go", "go"}}
	if got, want := workspaceText(f), "acme · main · 2 projects (2 Go)"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	f.Branch = ""
	if got, want := workspaceText(f), "acme · 2 projects (2 Go)"; got != want {
		t.Fatalf("no branch: got %q, want %q", got, want)
	}
}
