package termcolour

import (
	"strings"
	"testing"
)

func TestLineWriterFiltersWholeLinesAndFlushesTheLast(t *testing.T) {
	var out strings.Builder
	w := NewLineWriter(&out, true)
	_, _ = w.Write([]byte("--- PA"))
	_, _ = w.Write([]byte("SS: T (0s)\r\n\x1b[2Kdone"))
	w.Flush()
	if want := Wrap("--- PASS: T (0s)", Green) + "\ndone\n"; out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
