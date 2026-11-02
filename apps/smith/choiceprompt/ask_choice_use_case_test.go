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
