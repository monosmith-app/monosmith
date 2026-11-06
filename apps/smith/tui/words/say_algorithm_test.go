package words

import "testing"

func TestSayFillsPlaceholdersByName(t *testing.T) {
	if got := fill("{a} then {b}, {a} again", []any{"b", "two", "a", 1}); got != "1 then two, 1 again" {
		t.Fatalf("fill = %q", got)
	}
}

func TestAPlaceholderWithoutAValueStaysVisible(t *testing.T) {
	if got := fill("keep {missing}", []any{"other", "x"}); got != "keep {missing}" {
		t.Fatalf("fill = %q", got)
	}
}

func TestAKeyTheCatalogLacksSaysTheKey(t *testing.T) {
	if got := Say("no.such.key"); got != "no.such.key" {
		t.Fatalf("Say = %q", got)
	}
}

func TestCountPicksTheFormThatFitsN(t *testing.T) {
	catalogs["en"]["test.items.one"] = "{count} item"
	catalogs["en"]["test.items.other"] = "{count} items"
	defer delete(catalogs["en"], "test.items.one")
	defer delete(catalogs["en"], "test.items.other")
	for n, want := range map[int]string{1: "1 item", 0: "0 items", 3: "3 items"} {
		if got := Count("test.items", n); got != want {
			t.Errorf("Count(%d) = %q, want %q", n, got, want)
		}
	}
}
