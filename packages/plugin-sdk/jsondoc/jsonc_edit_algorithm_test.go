package jsondoc

import (
	"strings"
	"testing"
)

func refKeys(v any) map[string]any { return map[string]any{"references": v} }

func TestSetKeysInPlaceKeepsCommentsKeyOrderAndLayout(t *testing.T) {
	src := "{\n  // why\n  \"extends\": \"../../base.json\", /* base */\n  \"references\": [{ \"path\": \"../old\" }],\n  \"include\": [\"src\"],\n}\n"
	out, err := SetKeysInPlace([]byte(src), refKeys([]any{map[string]any{"path": "../core"}}))
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  // why\n  \"extends\": \"../../base.json\", /* base */\n  \"references\": [\n    {\n      \"path\": \"../core\"\n    }\n  ],\n  \"include\": [\"src\"],\n}\n"
	if string(out) != want {
		t.Errorf("got\n%s\nwant\n%s", out, want)
	}
}

func TestSetKeysInPlaceAddsAMissingKeyAtTheEnd(t *testing.T) {
	cases := map[string]string{
		"{\n    \"a\": 1\n}\n": "{\n    \"a\": 1,\n    \"references\": []\n}\n",
		"{}":                   "{\n  \"references\": []\n}",
	}
	for src, want := range cases {
		out, err := SetKeysInPlace([]byte(src), refKeys([]any{}))
		if err != nil || string(out) != want {
			t.Errorf("%q: got %q (%v), want %q", src, out, err, want)
		}
	}
}

func TestSetKeysInPlaceRefusesWhatIsNotAnObject(t *testing.T) {
	for _, src := range []string{"", "[]", "{\"a\" 1}", "{\"a\": 1"} {
		if _, err := SetKeysInPlace([]byte(src), map[string]any{"b": 1}); err == nil || !strings.Contains(err.Error(), "not a JSON object") {
			t.Errorf("%q: %v", src, err)
		}
	}
}

func TestStripCommentsKeepsStringsThatLookLikeComments(t *testing.T) {
	got := string(StripComments([]byte("{\"u\": \"http://x/*y*/\", // c\n \"a\": [1,],}")))
	if strings.Contains(got, "// c") || !strings.Contains(got, "http://x/*y*/") || strings.Contains(got, ",]") || strings.Contains(got, ",}") {
		t.Errorf("got %q", got)
	}
}
