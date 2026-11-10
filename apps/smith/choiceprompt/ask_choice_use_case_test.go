package choiceprompt

import (
	"bufio"
	"strings"
	"testing"
)

func TestAsk(t *testing.T) {
	opts := []Option{{"One", "a", "first"}, {"Two", "b", ""}}
	for _, tc := range []struct{ name, input, want string }{
		{"enter takes the default", "\n", "b"},
		{"a number", "1\n", "a"},
		{"a label in any case", "ONE\n", "a"},
		{"a value", "b\n", "b"},
		{"closed input takes the default", "", "b"},
		{"three wrong answers take the default", "x\ny\nz\n", "b"},
		{"a wrong answer then a right one", "9\n1\n", "a"},
	} {
		var out strings.Builder
		if got := Ask("Q?", opts, "b", bufio.NewReader(strings.NewReader(tc.input)), &out); got != tc.want {
			t.Errorf("%s: got %q, want %q\n%s", tc.name, got, tc.want, out.String())
		}
	}
}

func TestAskOrBackTakesBOnlyWhenAQuestionIsBefore(t *testing.T) {
	opts := []Option{{"One", "a", ""}, {"Back", "back", ""}}
	for _, tc := range []struct {
		name, input string
		canBack     bool
		want        string
		wantBack    bool
	}{
		{"b goes back", "b\n", true, "", true},
		{"an option named back keeps its meaning", "BACK\n", true, "back", false},
		{"b is no answer with no question before", "b\n1\n", false, "a", false},
	} {
		var out strings.Builder
		got, back := AskOrBack("Q?", opts, "a", bufio.NewReader(strings.NewReader(tc.input)), &out, tc.canBack)
		if got != tc.want || back != tc.wantBack {
			t.Errorf("%s: got %q back=%v, want %q back=%v\n%s", tc.name, got, back, tc.want, tc.wantBack, out.String())
		}
		if shown := strings.Contains(out.String(), "b) Back"); shown != tc.canBack {
			t.Errorf("%s: the b line shown=%v, want %v", tc.name, shown, tc.canBack)
		}
	}
}
