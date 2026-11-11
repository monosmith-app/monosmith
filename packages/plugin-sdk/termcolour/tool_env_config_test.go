package termcolour

import (
	"strings"
	"testing"
)

func TestEnvAsksForColourOrForPlainText(t *testing.T) {
	on, off := strings.Join(Env(true), " "), strings.Join(Env(false), " ")
	for _, want := range []string{"FORCE_COLOR=1", "CLICOLOR_FORCE=1", "CARGO_TERM_COLOR=always", "PY_COLORS=1"} {
		if !strings.Contains(on, want) {
			t.Errorf("colour env lacks %s", want)
		}
	}
	for _, want := range []string{"NO_COLOR=1", "FORCE_COLOR=0", "CARGO_TERM_COLOR=never", "PY_COLORS=0"} {
		if !strings.Contains(off, want) {
			t.Errorf("plain env lacks %s", want)
		}
	}
}
