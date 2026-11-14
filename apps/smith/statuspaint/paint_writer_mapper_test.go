package statuspaint

import (
	"bytes"
	"os"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

func TestWriterPaintsTheStatusLinesItGetsAndReportsTheLengthItWasGiven(t *testing.T) {
	var out bytes.Buffer
	w := writer{out: &out}
	text := "✓ the cache is within its limit\n"
	n, err := w.Write([]byte(text))
	if err != nil || n != len(text) {
		t.Fatalf("Write = %d, %v, want %d", n, err, len(text))
	}
	if want := termcolour.Wrap("✓", termcolour.Green) + " the cache is within its limit\n"; out.String() != want {
		t.Errorf("wrote %q, want %q", out.String(), want)
	}
}

func TestWriterReturnsTheDestinationItselfWhenColourIsNotWanted(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if got := Writer(os.Stdout); got != os.Stdout {
		t.Error("with NO_COLOR the writer must be the file itself, so a caller still sees a terminal")
	}
	var buf bytes.Buffer
	if got := Writer(&buf); got != &buf {
		t.Error("a buffer is not a terminal: no painting")
	}
}

func TestWantedFollowsTheTerminalDecisionAndNeverTheToolColourAHostHandsOn(t *testing.T) {
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	yes := func(int) bool { return true }
	var buf bytes.Buffer
	cases := []struct {
		name   string
		getenv func(string) string
		out    interface{ Write([]byte) (int, error) }
		want   bool
	}{
		{"a terminal", env(), os.Stdout, true},
		{"NO_COLOR", env("NO_COLOR", "1"), os.Stdout, false},
		{"a pipe", env(), &buf, false},
		{"FORCE_COLOR on a pipe", env("FORCE_COLOR", "1"), &buf, true},
		{"the interface child: tool colour on, FORCE_COLOR=0, a pipe", env(termcolour.ToolColourEnv, "1", "FORCE_COLOR", "0"), &buf, false},
	}
	for _, c := range cases {
		if got := Wanted(c.getenv, c.out, yes); got != c.want {
			t.Errorf("%s: Wanted = %v, want %v", c.name, got, c.want)
		}
	}
}
