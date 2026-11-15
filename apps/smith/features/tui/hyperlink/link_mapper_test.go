package hyperlink

import "testing"

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestLinkDrawsOsc8OnlyWhenOn(t *testing.T) {
	if got, want := Link("https://x.dev/a", "a", true), "\x1b]8;;https://x.dev/a\x1b\\a\x1b]8;;\x1b\\"; got != want {
		t.Errorf("link = %q", got)
	}
	if got := Link("https://x.dev/a", "a", false); got != "a" {
		t.Errorf("plain = %q", got)
	}
	if got := Link("", "a", true); got != "a" {
		t.Errorf("no address = %q", got)
	}
	if got := Link("https://x.dev/\x1b\\evil\x07", "a", true); got != "\x1b]8;;https://x.dev/\\evil\x1b\\a\x1b]8;;\x1b\\" {
		t.Errorf("control characters stay out of the sequence: %q", got)
	}
}

func TestSupportedFollowsTheTerminal(t *testing.T) {
	for _, c := range []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"windows terminal", map[string]string{"WT_SESSION": "x"}, true},
		{"iterm", map[string]string{"TERM_PROGRAM": "iTerm.app"}, true},
		{"vscode", map[string]string{"TERM_PROGRAM": "vscode"}, true},
		{"kitty", map[string]string{"TERM": "xterm-kitty"}, true},
		{"vte 0.50", map[string]string{"VTE_VERSION": "5000"}, true},
		{"old vte", map[string]string{"VTE_VERSION": "4800"}, false},
		{"unknown xterm", map[string]string{"TERM": "xterm-256color"}, false},
		{"apple terminal", map[string]string{"TERM_PROGRAM": "Apple_Terminal"}, false},
		{"tmux inside a known terminal", map[string]string{"WT_SESSION": "x", "TMUX": "/tmp/t"}, false},
		{"dumb", map[string]string{"TERM": "dumb", "WT_SESSION": "x"}, false},
		{"forced on", map[string]string{"FORCE_HYPERLINK": "1", "TERM": "dumb"}, true},
		{"forced off", map[string]string{"FORCE_HYPERLINK": "0", "WT_SESSION": "x"}, false},
	} {
		if got := Supported(env(c.env)); got != c.want {
			t.Errorf("%s: %v", c.name, got)
		}
	}
}
