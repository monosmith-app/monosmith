package languagenames

import "testing"

func TestEveryLanguageIsDescribedAndTheMostUsedWins(t *testing.T) {
	for id := range languages {
		if Describe(id) == "" {
			t.Errorf("%s has no description", id)
		}
	}
	if Describe("klingon") != "" {
		t.Fatal("unknown language described")
	}
	if got := MostUsed([]string{"go", "node", "node", "go", "rust"}); got != "go" {
		t.Fatalf("a tie goes to the first seen: %q", got)
	}
	if got := MostUsed([]string{"go", "node", "node"}); got != "node" {
		t.Fatalf("most used: %q", got)
	}
	if MostUsed(nil) != "" {
		t.Fatal("no projects, no default")
	}
}
