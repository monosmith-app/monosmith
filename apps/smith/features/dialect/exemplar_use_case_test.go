package dialect

import (
	"strings"
	"testing"
)

func TestExemplarFollowsTheChosenProject(t *testing.T) {
	var out strings.Builder
	d := Exemplar("core", map[string]string{"test.runner": "vitest", "layout": "src"}, "lib", map[string]bool{"test.runner": true}, &out)
	if d.Shape["test.runner"] != "vitest" || len(d.Shape) != 1 {
		t.Fatalf("shape %v", d.Shape)
	}
	if !strings.Contains(out.String(), "Following core") || !strings.Contains(out.String(), "layout = src (no plugin") {
		t.Errorf("output:\n%s", out.String())
	}
	if got := d.Warnings(map[string]string{"test.runner": "jest", "layout": "src"}); len(got) != 1 {
		t.Errorf("warnings %v", got)
	}
}
