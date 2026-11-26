package theme

import (
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

func TestTheLightThemeKeepsReadableContrast(t *testing.T) {
	const white = "#ffffff"
	l := Light
	for name, col := range map[string]string{"text": l.Text, "muted": l.Muted, "faint": l.Faint, "success": l.Success, "warning": l.Warning, "error": l.Error, "info": l.Info, "accent": l.Accent} {
		if c := hexContrast(t, col, white); c < 4.5 {
			t.Errorf("light %s %s has contrast %.2f on white, want 4.5", name, col, c)
		}
	}
	if c := hexContrast(t, l.Accent, l.Highlight); c < 4.5 {
		t.Errorf("the accent on the selected row has contrast %.2f, want 4.5", c)
	}
}

func TestTheDarkThemeKeepsReadableContrast(t *testing.T) {
	d := Dark
	for name, col := range map[string]string{"text": d.Text, "muted": d.Muted, "success": d.Success, "warning": d.Warning, "error": d.Error, "info": d.Info, "accent": d.Accent} {
		if c := hexContrast(t, col, "#000000"); c < 4.5 {
			t.Errorf("dark %s %s has contrast %.2f on black, want 4.5", name, col, c)
		}
	}
}

func TestASavedChoiceBeatsTheTerminalAndTheEnvironmentBeatsTheSavedChoice(t *testing.T) {
	cases := []struct {
		name         string
		env          map[string]string
		dark         bool
		theme, icons string
		wantTheme    string
		wantAscii    bool
	}{
		{"nothing saved follows the terminal", nil, true, "", "", "dark", false},
		{"auto is nothing saved", nil, false, "auto", "auto", "light", false},
		{"a saved theme beats the background", nil, true, "light", "", "light", false},
		{"saved icons", nil, true, "", "ascii", "dark", true},
		{"SMITH_THEME beats the saved theme", map[string]string{"SMITH_THEME": "plain"}, true, "light", "", "plain", false},
		{"NO_COLOR beats the saved theme", map[string]string{"NO_COLOR": "1"}, true, "dark", "", "plain", false},
		{"SMITH_ASCII beats saved unicode", map[string]string{"SMITH_ASCII": "1"}, true, "", "unicode", "dark", true},
		{"an unknown saved name is ignored", nil, true, "neon", "braille", "dark", false},
	}
	for _, c := range cases {
		got := ChooseWith(env(c.env), c.dark, c.theme, c.icons)
		if got.Tokens.Name != c.wantTheme || (got.Icons.Passed == ASCII.Passed) != c.wantAscii {
			t.Errorf("%s: %s, ascii=%v", c.name, got.Tokens.Name, got.Icons.Passed == ASCII.Passed)
		}
	}
}

func TestOverridesNameTheVariableThatWins(t *testing.T) {
	themeVar, iconsVar := Overrides(env(map[string]string{"SMITH_THEME": "Dark", "TERM": "dumb"}))
	if themeVar != "SMITH_THEME" || iconsVar != "TERM=dumb" {
		t.Fatalf("%q %q", themeVar, iconsVar)
	}
	if themeVar, iconsVar = Overrides(env(map[string]string{"SMITH_THEME": "neon"})); themeVar != "" || iconsVar != "" {
		t.Fatalf("an unknown theme overrides nothing: %q %q", themeVar, iconsVar)
	}
	if themeVar, _ = Overrides(env(map[string]string{"NO_COLOR": "1", "SMITH_THEME": "dark"})); themeVar != "NO_COLOR" {
		t.Fatalf("%q", themeVar)
	}
}
