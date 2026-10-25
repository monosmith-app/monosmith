package protocolschema

import "testing"

func TestParagraphsJoinLinesAndKeepParagraphs(t *testing.T) {
	got := Paragraphs("First line\n  continues here.\n\nSecond paragraph.\n")
	if want := "First line continues here.\n\nSecond paragraph."; got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if Paragraphs("") != "" {
		t.Error("no comment must be no text")
	}
}

func TestWithoutSubjectDropsTheGoName(t *testing.T) {
	cases := []struct{ name, doc, want string }{
		{"RoleLanguage", "RoleLanguage finds projects.", "Finds projects."},
		{"RoleLanguage", "Finds projects.", "Finds projects."},
		{"Role", "Role", "Role"},
	}
	for _, c := range cases {
		if got := WithoutSubject(c.name, c.doc); got != c.want {
			t.Errorf("%q: got %q want %q", c.doc, got, c.want)
		}
	}
}
