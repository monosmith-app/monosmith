package dialog

import (
	"strings"
	"testing"
)

func TestEmblemRowsAreASquareOfKnownColours(t *testing.T) {
	if len(emblemRows) != 32 {
		t.Fatalf("%d rows, want 32", len(emblemRows))
	}
	for y, row := range emblemRows {
		if len([]rune(row)) != 32 {
			t.Errorf("row %d has %d pixels, want 32", y, len([]rune(row)))
		}
		for _, ch := range row {
			if _, ok := emblemPalette[ch]; !ok && ch != '.' {
				t.Errorf("row %d: %q is not in the palette", y, ch)
			}
		}
	}
}

func TestEmblemGridKeepsTheCoreAndLeavesTheCornersOut(t *testing.T) {
	grid := emblemGrid()
	if grid[0][0] != nil {
		t.Error("a corner is a pixel")
	}
	if grid[16][16] == nil {
		t.Error("the centre is not a pixel")
	}
}

func TestShrinkHalvesTheGridAndDropsThinCorners(t *testing.T) {
	a, b := &rgb{200, 0, 0}, &rgb{100, 0, 0}
	got := shrink([][]*rgb{
		{a, b, nil, nil},
		{a, b, nil, a},
	})
	if len(got) != 1 || len(got[0]) != 2 {
		t.Fatalf("shrunk to %dx%d", len(got), len(got[0]))
	}
	if got[0][0] == nil || *got[0][0] != (rgb{150, 0, 0}) {
		t.Errorf("average = %v", got[0][0])
	}
	if got[0][1] != nil {
		t.Errorf("one pixel of four stayed: %v", got[0][1])
	}
}

func TestHalfBlocksDrawTwoPixelRowsToATextRow(t *testing.T) {
	px := &rgb{1, 2, 3}
	paint := func(glyph string, fg, bg *rgb) string {
		if bg != nil {
			return "[" + glyph + "+bg]"
		}
		return "[" + glyph + "]"
	}
	lines := halfBlocks([][]*rgb{
		{nil, px, nil, px},
		{nil, nil, px, px},
	}, paint)
	if len(lines) != 1 {
		t.Fatalf("%d lines", len(lines))
	}
	if want := " [▀][▄][▀+bg]"; lines[0] != want {
		t.Errorf("%q, want %q", lines[0], want)
	}
}

func TestEmblemHexIsLowerCaseAndPadded(t *testing.T) {
	if got := (rgb{0x70, 0x20, 0x0d}).hex(); got != "#70200d" {
		t.Errorf("hex = %s", got)
	}
}

func TestTheEmblemFitsTheDialog(t *testing.T) {
	lines := halfBlocks(emblemGrid(), func(glyph string, _, _ *rgb) string { return glyph })
	if len(lines) != 16 {
		t.Errorf("%d text rows, want 16", len(lines))
	}
	for _, l := range lines {
		if n := len([]rune(l)); n != 32 {
			t.Errorf("a row is %d columns wide: %q", n, strings.TrimSpace(l))
		}
	}
}

func TestEmblemSizeFollowsTheTerminalHeight(t *testing.T) {
	for height, want := range map[int]emblemSize{24: emblemAnvil, 33: emblemAnvil, 34: emblemSmall, 43: emblemSmall, 44: emblemLarge, 80: emblemLarge} {
		if got := emblemSizeFor(height); got != want {
			t.Errorf("height %d: size %d, want %d", height, got, want)
		}
	}
}
