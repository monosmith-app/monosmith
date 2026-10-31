package languageserver

import (
	"strings"
	"testing"
)

func TestHoverOnKeyShowsPathTypeAndDescription(t *testing.T) {
	h := HoverAt("[workspace]\nname = \"a\"\n", Position{1, 2})
	if h == nil {
		t.Fatal("no hover")
	}
	v := h.Contents.Value
	if h.Contents.Kind != "markdown" || !strings.Contains(v, "workspace.name") || !strings.Contains(v, "(string)") || !strings.Contains(v, "workspace's name") {
		t.Errorf("value %q", v)
	}
}

func TestHoverOnTableHeader(t *testing.T) {
	h := HoverAt("[ci]\n", Position{0, 2})
	if h == nil || !strings.Contains(h.Contents.Value, "**ci**") {
		t.Fatalf("got %+v", h)
	}
}

func TestHoverOnDottedKeyUsesTheSegmentUnderTheCursor(t *testing.T) {
	h := HoverAt("workspace.name = \"a\"\n", Position{0, 3})
	if h == nil || !strings.Contains(h.Contents.Value, "**workspace**") {
		t.Fatalf("got %+v", h)
	}
}

func TestHoverIsNilOffTheSchema(t *testing.T) {
	for _, c := range []struct {
		text string
		pos  Position
	}{
		{"[workspace]\nname = \"abc\"\n", Position{1, 10}},
		{"[workspace]\nnope = 1\n", Position{1, 1}},
		{"[workspace]\nname = \"a\"\n", Position{1, 6}},
		{"[workspace]\n", Position{9, 0}},
	} {
		if h := HoverAt(c.text, c.pos); h != nil {
			t.Errorf("%q at %+v: want nil, got %+v", c.text, c.pos, h)
		}
	}
}
