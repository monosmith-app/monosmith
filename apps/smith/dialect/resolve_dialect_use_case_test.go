package dialect

import (
	"slices"
	"strings"
	"testing"
)

func lib(runner string) map[string]string {
	return map[string]string{"test.runner": runner, "build.tool": "esbuild", "layout": "src"}
}

func TestResolveFollowsWhatTheSiblingsAgreeOnAndSaysWhy(t *testing.T) {
	var out strings.Builder
	siblings := []map[string]string{lib("@swc/jest"), lib("@swc/jest"), lib("@swc/jest")}
	d := Resolve(Request{Siblings: siblings, Kind: "lib", Follows: map[string]bool{"test.runner": true, "build.tool": true}, Out: &out})
	if d.Shape["test.runner"] != "@swc/jest" || d.Shape["build.tool"] != "esbuild" {
		t.Fatalf("shape %v", d.Shape)
	}
	if _, sent := d.Shape["layout"]; sent {
		t.Error("a key no plugin follows is not sent")
	}
	if !strings.Contains(out.String(), "test.runner = @swc/jest (3 of 3 libs)") {
		t.Errorf("the reason is missing:\n%s", out.String())
	}
	if got := d.Warnings(map[string]string{"test.runner": "@swc/jest", "build.tool": "esbuild", "layout": "src"}); len(got) != 0 {
		t.Errorf("a matching project has no warning: %v", got)
	}
	got := d.Warnings(map[string]string{"test.runner": "jest", "build.tool": "esbuild", "layout": "src,test"})
	if len(got) != 1 || !strings.Contains(got[0], "3 of 3 sibling libs have test.runner = @swc/jest, the new lib has jest") || !strings.Contains(got[0], "layout = src, the new lib has src,test") {
		t.Errorf("warnings %v", got)
	}
}

func TestResolveSaysNothingWithFewerThanThreeSiblings(t *testing.T) {
	var out strings.Builder
	d := Resolve(Request{Siblings: []map[string]string{lib("jest"), lib("jest")}, Kind: "lib", Follows: map[string]bool{"test.runner": true}, Out: &out})
	if d.Shape != nil || out.Len() != 0 {
		t.Fatalf("shape %v, output %q", d.Shape, out.String())
	}
}

func split() []map[string]string {
	return []map[string]string{lib("jest"), lib("jest"), lib("vitest"), lib("vitest")}
}

func TestResolveAsksOnASplitAndNeverGuesses(t *testing.T) {
	var out strings.Builder
	d := Resolve(Request{Siblings: split(), Kind: "lib", Follows: map[string]bool{"test.runner": true}, Interactive: true, In: strings.NewReader("2\n"), Out: &out})
	if d.Shape["test.runner"] != "vitest" {
		t.Fatalf("the answer is followed: %v", d.Shape)
	}
	for _, want := range []string{"split on test.runner", "1) jest (2 libs)", "2) vitest (2 libs)", "3) the default of the stack"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the question lacks %q:\n%s", want, out.String())
		}
	}
}

func TestResolveAnswerOfTheDefaultFollowsNothingForThatKey(t *testing.T) {
	for _, answer := range []string{"3\n", "\n", ""} {
		d := Resolve(Request{Siblings: split(), Kind: "lib", Follows: map[string]bool{"test.runner": true}, Interactive: true, In: strings.NewReader(answer), Out: &strings.Builder{}})
		if _, followed := d.Shape["test.runner"]; followed {
			t.Errorf("answer %q follows %v", answer, d.Shape)
		}
	}
}

func TestResolveNeverBlocksWithoutATerminal(t *testing.T) {
	var out strings.Builder
	d := Resolve(Request{Siblings: split(), Kind: "lib", Follows: map[string]bool{"test.runner": true}, Interactive: false, In: nil, Out: &out})
	if _, followed := d.Shape["test.runner"]; followed {
		t.Errorf("a split is not guessed: %v", d.Shape)
	}
	if !strings.Contains(out.String(), "Warning: the sibling libs split on test.runner (jest 2, vitest 2); using the default") {
		t.Errorf("the split is not warned about:\n%s", out.String())
	}
}

func TestResolveOnlyReportsASplitNoPluginFollows(t *testing.T) {
	var out strings.Builder
	Resolve(Request{Siblings: split(), Kind: "lib", Interactive: true, In: strings.NewReader("1\n"), Out: &out})
	if strings.Contains(out.String(), "Which should") || !strings.Contains(out.String(), "differ on test.runner") {
		t.Errorf("output:\n%s", out.String())
	}
	if !slices.Contains(strings.Split(out.String(), "\n"), "The sibling libs differ on test.runner: jest 2, vitest 2") {
		t.Errorf("output:\n%s", out.String())
	}
}
