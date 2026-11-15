package theme

import (
	"math"
	"strconv"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// hexLuminance is the relative luminance of a #rrggbb colour (WCAG 2).
func hexLuminance(t *testing.T, hex string) float64 {
	t.Helper()
	var c [3]float64
	for i := range c {
		v, err := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		if err != nil {
			t.Fatalf("%q is not #rrggbb", hex)
		}
		s := float64(v) / 255
		if s <= 0.03928 {
			c[i] = s / 12.92
		} else {
			c[i] = math.Pow((s+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*c[0] + 0.7152*c[1] + 0.0722*c[2]
}

// hexContrast is the WCAG contrast ratio of two colours.
func hexContrast(t *testing.T, a, b string) float64 {
	la, lb := hexLuminance(t, a), hexLuminance(t, b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestTheButtonHoverColourIsBetweenRestAndPressed(t *testing.T) {
	for _, tok := range []Tokens{Dark, Light} {
		rest, hover, active := hexLuminance(t, tok.Button), hexLuminance(t, tok.ButtonHover), hexLuminance(t, tok.ButtonActive)
		lo, hi := min(rest, active), max(rest, active)
		if hover <= lo || hover >= hi {
			t.Errorf("%s: hover %.3f is not between rest %.3f and pressed %.3f", tok.Name, hover, rest, active)
		}
		if c := hexContrast(t, tok.ButtonHover, tok.ButtonText); c < 4.5 {
			t.Errorf("%s: the text on a hovered button has a contrast of %.1f, under 4.5", tok.Name, c)
		}
	}
}

func TestTheRowHoverIsBetweenTheBackgroundAndTheSelectedRow(t *testing.T) {
	// The terminal's own background: black (16) for dark, white (231) for light.
	for tok, bg := range map[*Tokens]int{&Dark: 16, &Light: 231} {
		hover, err1 := strconv.Atoi(tok.RowHover)
		selected, err2 := strconv.Atoi(tok.Highlight)
		if err1 != nil || err2 != nil {
			t.Fatalf("%s: row hover %q and highlight %q are 256-colour numbers", tok.Name, tok.RowHover, tok.Highlight)
		}
		lo, hi := min(luminance(bg), luminance(selected)), max(luminance(bg), luminance(selected))
		if l := luminance(hover); l <= lo || l >= hi {
			t.Errorf("%s: row hover %d is not between the background %d and the selected row %d", tok.Name, hover, bg, selected)
		}
	}
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	if New(Choice{Tokens: Plain, Icons: ASCII}).RowHover("row") == "row" {
		t.Error("without colour a hovered row is still marked (underlined)")
	}
}
