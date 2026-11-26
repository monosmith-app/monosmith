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

func TestEveryTextColourKeepsAContrastOf4Point5OnEveryBackground(t *testing.T) {
	for tok, terminal := range map[*Tokens]string{&Dark: "#000000", &Light: "#ffffff"} {
		texts := map[string]string{"accent": tok.Accent, "focus": tok.Focus, "info": tok.Info, "success": tok.Success,
			"warning": tok.Warning, "error": tok.Error, "text": tok.Text, "muted": tok.Muted, "faint": tok.Faint}
		backgrounds := map[string]string{"terminal": terminal, "highlight": tok.Highlight, "bar": tok.Bar, "row hover": tok.RowHover}
		for tn, fg := range texts {
			for bn, bg := range backgrounds {
				if c := hexContrast(t, fg, bg); c < 4.5 {
					t.Errorf("%s: %s %s on %s %s has a contrast of %.2f, under 4.5", tok.Name, tn, fg, bn, bg, c)
				}
			}
		}
		buttons := map[string][2]string{"primary": {tok.ButtonOn, tok.ButtonActive}, "secondary": {tok.ButtonText, tok.Button}}
		for name, pair := range buttons {
			if c := hexContrast(t, pair[0], pair[1]); c < 4.5 {
				t.Errorf("%s: the %s button has a contrast of %.2f, under 4.5", tok.Name, name, c)
			}
		}
	}
}

func TestThePaletteIsTheMonoSmithOne(t *testing.T) {
	for _, c := range []struct{ got, want string }{
		{Dark.Accent, "#5fb6d1"}, {Light.Accent, "#176a86"}, {Dark.Focus, Dark.Accent}, {Light.Info, Light.Accent},
		{Dark.Text, "#d0d0d0"}, {Light.Muted, "#5c5c5c"}, {Dark.Highlight, "#2f3d44"}, {Light.Bar, "#f2f2f2"},
		{Dark.Button, Dark.Highlight}, {Dark.ButtonText, Dark.Text}, {Dark.ButtonActive, Dark.Accent}, {Dark.ButtonOn, "#141414"},
		{Light.Button, Light.Highlight}, {Light.ButtonText, Light.Text}, {Light.ButtonActive, Light.Accent}, {Light.ButtonOn, "#ffffff"},
	} {
		if c.got != c.want {
			t.Errorf("token %s, want %s", c.got, c.want)
		}
	}
	if Plain != (Tokens{Name: "plain"}) {
		t.Errorf("the plain theme has no colour: %+v", Plain)
	}
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
	// The terminal's own background: black for dark, white for light.
	for tok, bg := range map[*Tokens]string{&Dark: "#000000", &Light: "#ffffff"} {
		lb, ls := hexLuminance(t, bg), hexLuminance(t, tok.Highlight)
		lo, hi := min(lb, ls), max(lb, ls)
		if l := hexLuminance(t, tok.RowHover); l <= lo || l >= hi {
			t.Errorf("%s: row hover %s is not between the background %s and the selected row %s", tok.Name, tok.RowHover, bg, tok.Highlight)
		}
	}
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.TrueColor)
	if New(Choice{Tokens: Plain, Icons: ASCII}).RowHover("row") == "row" {
		t.Error("without colour a hovered row is still marked (underlined)")
	}
}
