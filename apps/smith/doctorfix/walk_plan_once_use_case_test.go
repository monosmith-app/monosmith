package doctorfix

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestWalkShowsAFixWhoseQuestionsAreAnsweredUpFrontOnce(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	plan := planFor(t, root)[1:]
	if _, err := Walk(context.Background(), plan, Options{Root: root, Out: &out, Yes: true, Explicit: true, Choose: []int{2}, Command: command}); err != nil {
		t.Fatal(err)
	}
	// The fix is shown once ("your choices: ..."), then reported applied.
	if n := strings.Count(out.String(), "oxfmt is missing"); n != 2 {
		t.Fatalf("the description appears %d times, want 2:\n%s", n, out.String())
	}
}
