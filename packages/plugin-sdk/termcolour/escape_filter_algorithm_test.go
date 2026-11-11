package termcolour

import (
	"strings"
	"testing"
)

func TestKeepSGRDropsCursorMovementAndProgressButKeepsColour(t *testing.T) {
	in := "\x1b[?25l\x1b[2K\x1b[1G\x1b[32mok\x1b[0m \x1b]0;title\x07\x1b]8;;http://x\x1b\\link\x1b]8;;\x1b\\\x1b7\x1b[38;5;196mred\x1b[m\x1b[1A"
	want := "\x1b[32mok\x1b[0m link\x1b[38;5;196mred\x1b[m"
	if got := KeepSGR(in); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := Strip(in); got != "ok linkred" {
		t.Errorf("Strip: got %q", got)
	}
	if strings.Contains(Strip("\x1b[31mhalf"), "\x1b") || Strip("\x1b[") != "" {
		t.Error("an unfinished sequence must not survive Strip")
	}
}
