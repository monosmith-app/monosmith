package languageserver

import "testing"

func labels(items []CompletionItem) map[string]CompletionItem {
	m := map[string]CompletionItem{}
	for _, i := range items {
		m[i.Label] = i
	}
	return m
}

func TestCompleteTableNamesAfterBracket(t *testing.T) {
	got := labels(Complete("[w", Position{0, 2}))
	if _, ok := got["workspace"]; !ok || len(got) != 1 {
		t.Fatalf("got %v", got)
	}
	if got["workspace"].Kind != CompletionKindModule {
		t.Errorf("kind %d", got["workspace"].Kind)
	}
}

func TestCompleteKeysOfTheCurrentTable(t *testing.T) {
	text := "[workspace]\nname = \"a\"\ns"
	got := labels(Complete(text, Position{2, 1}))
	scope, ok := got["scope"]
	// "s" also matches shared_outputs, a [workspace] key since the outputs rules.
	if _, shared := got["shared_outputs"]; !ok || !shared || len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if scope.InsertText != "scope = " || scope.Documentation == "" || scope.Kind != CompletionKindProperty {
		t.Errorf("item %+v", scope)
	}
}

func TestCompleteTopLevelKeysOffersSchema(t *testing.T) {
	got := labels(Complete("sch", Position{0, 3}))
	if _, ok := got["schema"]; !ok {
		t.Fatalf("got %v", got)
	}
}

func TestCompleteEnumValuesAreQuoted(t *testing.T) {
	text := "[ci]\nproviders = ["
	got := labels(Complete(text, Position{1, 13}))
	if got["github"].InsertText != `"github"` || len(got) != 5 {
		t.Fatalf("got %v", got)
	}
}

func TestCompleteBooleanValues(t *testing.T) {
	got := labels(Complete("[ci]\ndevcontainer = ", Position{1, 15}))
	if got["true"].InsertText != "true" || got["false"].Label != "false" {
		t.Fatalf("got %v", got)
	}
}

func TestCompleteNothingForUnknownContext(t *testing.T) {
	for _, c := range []struct {
		text string
		pos  Position
	}{
		{"[nope]\nx", Position{1, 1}},
		{"# c", Position{0, 3}},
		{"[ci]\nname = ", Position{1, 7}},
	} {
		if got := Complete(c.text, c.pos); len(got) != 0 || got == nil {
			t.Errorf("%q: want empty non-nil, got %v", c.text, got)
		}
	}
}
