package statuspaint

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

func TestPaintColoursOnlyTheGlyphOfAStatusLine(t *testing.T) {
	cases := map[string]string{
		"kit:build ✓ 106ms":                      "kit:build " + termcolour.Wrap("✓", termcolour.Green) + " 106ms",
		"orders:build × exited with code 2":      "orders:build " + termcolour.Wrap("×", termcolour.Red) + " exited with code 2",
		"  × orders:build exited with code 2":    "  " + termcolour.Wrap("×", termcolour.Red) + " orders:build exited with code 2",
		"  » web:build because orders failed":    "  " + termcolour.Wrap("»", termcolour.Dim) + " web:build because orders failed",
		"✓ 4/4 tasks succeeded (0 from cache)":   termcolour.Wrap("✓", termcolour.Green) + " 4/4 tasks succeeded (0 from cache)",
		"kit:build ▶ go build ./...":             "kit:build " + termcolour.Wrap("▶", termcolour.Dim) + " go build ./...",
		"kit:build ! not cached: disk full":      "kit:build " + termcolour.Wrap("!", termcolour.Yellow) + " not cached: disk full",
		"kit:build → run smith install":          "kit:build " + termcolour.Wrap("→", termcolour.Dim) + " run smith install",
		"kit:build │ ✓ a tool prints this":       "kit:build │ ✓ a tool prints this",
		"just a sentence with a ✓ in the middle": "just a sentence with a ✓ in the middle",
	}
	for in, want := range cases {
		if got := Paint(in); got != want {
			t.Errorf("Paint(%q)\n got %q\nwant %q", in, got, want)
		}
		if strings.Contains(termcolour.Strip(Paint(in)), "\x1b") || termcolour.Strip(Paint(in)) != in {
			t.Errorf("painting changed the characters of %q", in)
		}
	}
}

func TestPaintLeavesALineThatIsColouredAlone(t *testing.T) {
	line := "kit:build ✓ " + termcolour.Wrap("ok", termcolour.Green)
	if got := Paint(line); got != line {
		t.Errorf("a coloured line was painted again: %q", got)
	}
}
