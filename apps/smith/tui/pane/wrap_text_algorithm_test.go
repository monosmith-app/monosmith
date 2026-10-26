package pane

import (
	"strings"
	"testing"
)

func TestAPathWrapsAtItsSlashesNotInTheMiddleOfAName(t *testing.T) {
	lines := Wrap("Import it as example.com/jato.auto.packages/libs/geo-utils/something", 40, 0)
	if len(lines) < 2 {
		t.Fatalf("not wrapped: %q", lines)
	}
	for _, l := range lines[:len(lines)-1] {
		if !strings.HasSuffix(l, "/") && !strings.HasSuffix(l, " ") && !strings.HasSuffix(l, "-") {
			t.Fatalf("line %q breaks inside a name (all: %q)", l, lines)
		}
	}
}
