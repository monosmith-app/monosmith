package theme

import (
	"math"
	"strconv"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestChooseHonoursNoColorThePlainAndLightThemesAndTheBackground(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		dark bool
		want string
	}{
		{"a dark terminal", nil, true, "dark"},
		{"a light terminal is detected", nil, false, "light"},
		{"SMITH_THEME=light forces light", map[string]string{"SMITH_THEME": "light"}, true, "light"},
		{"SMITH_THEME=dark forces dark", map[string]string{"SMITH_THEME": "dark"}, false, "dark"},
		{"SMITH_THEME=plain", map[string]string{"SMITH_THEME": "plain"}, true, "plain"},
		{"NO_COLOR wins over the theme", map[string]string{"NO_COLOR": "1", "SMITH_THEME": "light"}, true, "plain"},
	}
	for _, c := range cases {
		if got := Choose(env(c.env), c.dark).Tokens.Name; got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestTheLightBackgroundIsDetectedFromColorFgBg(t *testing.T) {
	if DarkBackground(env(map[string]string{"COLORFGBG": "0;15"})) {
		t.Error("0;15 is dark text on a light background")
	}
	if !DarkBackground(env(map[string]string{"COLORFGBG": "15;0"})) {
		t.Error("15;0 is light text on a dark background")
	}
}

func TestEveryStateHasASymbolSoItSurvivesWithoutColour(t *testing.T) {
	for _, set := range []Icons{Unicode, ASCII} {
		for name, s := range map[string]string{"passed": set.Passed, "failed": set.Failed, "warning": set.Warning, "running": set.Running} {
			if s == "" {
				t.Errorf("%s has no symbol", name)
			}
		}
		if set.Passed == set.Failed || set.Failed == set.Warning || set.Passed == set.Warning {
			t.Errorf("passed, failed and warning must differ: %+v", set)
		}
	}
}

func TestPlainTextCarriesNoEscapeCodes(t *testing.T) {
	th := New(Choice{Tokens: Plain, Icons: Unicode})
	if got := th.Success("ok"); got != "ok" {
		t.Errorf("plain success is %q", got)
	}
}

// ansi256 is the sRGB colour of an xterm 256-colour number (16..255).
func ansi256(n int) (r, g, b float64) {
	if n >= 232 {
		v := float64(8+10*(n-232)) / 255
		return v, v, v
	}
	n -= 16
	level := func(i int) float64 {
		if i == 0 {
			return 0
		}
		return float64(55+40*i) / 255
	}
	return level(n / 36), level(n / 6 % 6), level(n % 6)
}

func luminance(n int) float64 {
	lin := func(c float64) float64 {
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	r, g, b := ansi256(n)
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

func contrast(a, b int) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestTheLightThemeKeepsReadableContrast(t *testing.T) {
	const white = 231
	l := Light
	for name, col := range map[string]string{"text": l.Text, "muted": l.Muted, "faint": l.Faint, "success": l.Success, "warning": l.Warning, "error": l.Error, "info": l.Info, "accent": l.Accent} {
		n, err := strconv.Atoi(col)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if c := contrast(n, white); c < 4.5 {
			t.Errorf("light %s %s has contrast %.2f on white, want 4.5", name, col, c)
		}
	}
	acc, _ := strconv.Atoi(l.Accent)
	hl, _ := strconv.Atoi(l.Highlight)
	if c := contrast(acc, hl); c < 4.5 {
		t.Errorf("the accent on the selected row has contrast %.2f, want 4.5", c)
	}
}

func TestTheDarkThemeKeepsReadableContrast(t *testing.T) {
	d := Dark
	for name, col := range map[string]string{"text": d.Text, "muted": d.Muted, "success": d.Success, "warning": d.Warning, "error": d.Error, "info": d.Info, "accent": d.Accent} {
		n, _ := strconv.Atoi(col)
		bg, _ := strconv.Atoi("16")
		if c := contrast(n, bg); c < 4.5 {
			t.Errorf("dark %s %s has contrast %.2f on black, want 4.5", name, col, c)
		}
	}
}
