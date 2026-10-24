package screenreader

import "testing"

func TestWantedFollowsTheEnvironmentVariable(t *testing.T) {
	cases := map[string]bool{"": false, "0": false, "false": false, "No": false, "off": false, "1": true, "true": true, "yes": true, "anything": true}
	for value, want := range cases {
		got := Wanted(func(name string) string {
			if name != Variable {
				t.Fatalf("looked at %q", name)
			}
			return value
		})
		if got != want {
			t.Errorf("%s=%q: %v, want %v", Variable, value, got, want)
		}
	}
}

func TestPlainWhenAskedWhenThemeIsPlainOrWhenOutputIsNotATerminal(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	cases := []struct {
		name string
		env  map[string]string
		tty  bool
		want bool
	}{
		{"a terminal and no setting keeps the shell", nil, true, false},
		{"output that is not a terminal", nil, false, true},
		{"the screen reader variable", map[string]string{Variable: "1"}, true, true},
		{"the plain theme", map[string]string{ThemeVariable: "Plain"}, true, true},
		{"NO_COLOR alone keeps the shell", map[string]string{"NO_COLOR": "1"}, true, false},
		{"the light theme keeps the shell", map[string]string{ThemeVariable: "light"}, true, false},
	}
	for _, c := range cases {
		if got := Plain(env(c.env), c.tty); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}
