package recipechoice

import (
	"strings"
	"testing"
)

func TestParseAnswers(t *testing.T) {
	got, err := ParseAnswers([]string{"language=python", " kind = library "})
	if err != nil || got["language"] != "python" || got["kind"] != "library" || len(got) != 2 {
		t.Fatalf("got %v, %v", got, err)
	}
	if got, err := ParseAnswers(nil); err != nil || len(got) != 0 {
		t.Fatalf("none: %v, %v", got, err)
	}
	for in, want := range map[string]string{
		"language":  "name=answer",
		"=python":   "name=answer",
		"language=": "name=answer",
	} {
		if _, err := ParseAnswers([]string{in}); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: want %q, got %v", in, want, err)
		}
	}
	if _, err := ParseAnswers([]string{"a=1", "a=2"}); err == nil || !strings.Contains(err.Error(), "given twice") {
		t.Errorf("duplicate: %v", err)
	}
}
