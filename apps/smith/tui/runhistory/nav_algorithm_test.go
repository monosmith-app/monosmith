package runhistory

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNavPartsDisableWhatCannotBeUsed(t *testing.T) {
	on := func(parts []NavPart, action string) bool {
		for _, p := range parts {
			if p.Action == action {
				return p.On
			}
		}
		t.Fatalf("no part %q", action)
		return false
	}
	first := NavParts(true, 0, 3)
	if on(first, "older") || !on(first, "newer") || !on(first, "start") {
		t.Fatalf("first of three: %+v", first)
	}
	last := NavParts(true, 2, 3)
	if !on(last, "older") || on(last, "newer") {
		t.Fatalf("last of three: %+v", last)
	}
	busy := NavParts(false, 1, 3)
	if on(busy, "start") || on(busy, "older") || on(busy, "newer") {
		t.Fatalf("a run goes: %+v", busy)
	}
}

func TestActionAtFindsThePartUnderAColumn(t *testing.T) {
	parts := NavParts(true, 1, 3)
	var line strings.Builder
	for _, p := range parts {
		line.WriteString(p.Text)
	}
	if got := ActionAt(parts, 1); got != "start" {
		t.Fatalf("column 1 is %q", got)
	}
	older := utf8.RuneCountInString(line.String()[:strings.Index(line.String(), "older")])
	if got := ActionAt(parts, older+1); got != "older" {
		t.Fatalf("column on older is %q", got)
	}
	if got := ActionAt(parts, 500); got != "" {
		t.Fatalf("past the end is %q", got)
	}
}
