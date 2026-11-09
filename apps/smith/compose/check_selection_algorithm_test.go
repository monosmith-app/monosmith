package compose

import (
	"strings"
	"testing"
)

func TestACompleteSelectionAddsTheDefaultsAndSaysWhich(t *testing.T) {
	c := CheckPicks(fakeOffers(), "", []string{"react"})
	if !c.Complete || c.Language != "node" {
		t.Fatalf("%+v", c)
	}
	if strings.Join(c.Selection, ",") != "react,eslint,vitest" || strings.Join(c.Added, ",") != "eslint,vitest" {
		t.Fatalf("selection %v added %v", c.Selection, c.Added)
	}
}

func TestAConflictNamesBothItemsAndTheFix(t *testing.T) {
	c := CheckPicks(fakeOffers(), "node", []string{"react", "electron"})
	for _, want := range []string{"react", "electron", "framework axis", "keep one"} {
		if !strings.Contains(c.Conflict, want) {
			t.Fatalf("the conflict %q lacks %q", c.Conflict, want)
		}
	}
	if c.Complete {
		t.Fatal("a conflict is not a complete selection")
	}
}

func TestADeclaredConflictNamesBothItems(t *testing.T) {
	offers := fakeOffers()
	c := CheckPicks(offers, "node", []string{"electron", "bootstrap"})
	if c.Conflict != "" {
		t.Fatalf("a style does not conflict with a shell: %q", c.Conflict)
	}
	c = CheckPicks(offers, "node", []string{"electron", "azure-functions"})
	if !strings.Contains(c.Conflict, "electron") || !strings.Contains(c.Conflict, "azure-functions") {
		t.Fatalf("electron with a function host is explained: %+v", c)
	}
}

func TestAnOpenAxisBecomesAQuestionWithTheSuggestedAnswerMarked(t *testing.T) {
	c := CheckPicks(fakeOffers(), "node", []string{"azure-functions"})
	if c.Question == nil || c.Question.Axis != "trigger" || c.Conflict != "" || c.Complete {
		t.Fatalf("%+v", c)
	}
	var suggested string
	for _, o := range c.Question.Options {
		if o.Suggested {
			suggested = o.Name
		}
	}
	if suggested != "http-trigger" || len(c.Question.Options) != 2 {
		t.Fatalf("options %+v", c.Question.Options)
	}
	answered := CheckPicks(fakeOffers(), "node", []string{"azure-functions", "http-trigger"})
	if !answered.Complete {
		t.Fatalf("the answer completes it: %+v", answered)
	}
}

func TestASelectionWithNoStarterIsMissingNotWrong(t *testing.T) {
	c := CheckPicks(fakeOffers(), "node", []string{"bootstrap"})
	if c.Conflict != "" || c.Missing == "" || c.Complete {
		t.Fatalf("a style alone is not an error, it needs a starter: %+v", c)
	}
	if empty := CheckPicks(fakeOffers(), "", nil); empty.Missing == "" || empty.Conflict != "" {
		t.Fatalf("nothing picked asks for the first pick: %+v", empty)
	}
}

func TestTheLanguageComesFromThePicksAndAMixIsExplained(t *testing.T) {
	cat := Build(Source{Offers: fakeOffers(), Languages: []Language{{Name: "node"}, {Name: "python"}}})
	react, _ := cat.Find("node/react")
	py, _ := cat.Find("python/azure-functions")
	c := CheckItems(OffersOf(cat.Items), "", []Item{react, py})
	if !strings.Contains(c.Conflict, "python") || !strings.Contains(c.Conflict, "node") {
		t.Fatalf("a node starter with a python host is a conflict that names both languages: %+v", c)
	}
	if ok := CheckItems(OffersOf(cat.Items), "", []Item{py}); ok.Language != "python" || ok.Conflict != "" {
		t.Fatalf("one pick fixes the language: %+v", ok)
	}
}
