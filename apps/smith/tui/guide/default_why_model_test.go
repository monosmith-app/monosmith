package guide

import (
	"strings"
	"testing"
)

func TestTheLanguageQuestionSaysWhyItStartsOnItsDefault(t *testing.T) {
	m := menu().WithDefault("lang", "go").WithDefaultWhy("lang", "Preselected: 2 of the 4 projects here use Go.")
	m = keys(m, "down", "enter", "enter").(Model)
	if !strings.Contains(m.View(), "Preselected: 2 of the 4 projects here use Go.") {
		t.Fatalf("the reason is shown:\n%s", m.View())
	}
	plain := keys(menu(), "down", "enter", "enter").(Model)
	if strings.Contains(plain.View(), "Preselected") {
		t.Fatalf("no reason without a default:\n%s", plain.View())
	}
}

func TestABadNameIsToldOnceNotTwice(t *testing.T) {
	m := keys(menu(), "down", "enter", "down", "down", "enter", "enter", "B", "a", "d", "enter")
	if n := strings.Count(m.View(), "lowercase words joined by dashes, like"); n != 1 {
		t.Fatalf("the hint appears %d times:\n%s", n, m.View())
	}
}
