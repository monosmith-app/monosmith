package languageserver

import (
	"reflect"
	"testing"
)

func TestSplitKey(t *testing.T) {
	for in, want := range map[string][]string{
		"":                      {""},
		"a":                     {"a"},
		"a.b":                   {"a", "b"},
		` projects . "my.app" `: {"projects", "my.app"},
		"a.":                    {"a", ""},
	} {
		if got := splitKey(in); !reflect.DeepEqual(got, want) {
			t.Errorf("splitKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLocateKinds(t *testing.T) {
	text := "[workspace]\nname = \"a\"\n\n[ci]\nprov\n"
	cases := []struct {
		name string
		pos  Position
		want cursor
	}{
		{"key in table", Position{4, 4}, cursor{kind: cursorKey, path: []string{"ci", "prov"}, prefix: "prov"}},
		{"value", Position{1, 7}, cursor{kind: cursorValue, path: []string{"workspace", "name"}, prefix: ""}},
		{"header", Position{3, 3}, cursor{kind: cursorHeader, path: []string{"ci"}, prefix: "ci"}},
		{"after the closing bracket", Position{3, 4}, cursor{}},
		{"out of range", Position{99, 0}, cursor{}},
	}
	for _, c := range cases {
		if got := locate(text, c.pos); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestLocateIgnoresCommentsAndArrayLines(t *testing.T) {
	if got := locate("# nam", Position{0, 5}); got.kind != cursorNone {
		t.Errorf("comment: %+v", got)
	}
	text := "[ci]\nbranches = [\n  [\"a\"],\n]\nx"
	if got := locate(text, Position{4, 1}); !reflect.DeepEqual(got.path, []string{"ci", "x"}) {
		t.Errorf("array line took for a header: %+v", got)
	}
}
