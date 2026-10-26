package theme

import (
	"slices"
	"strings"
)

// Choice says which tokens and icons to use.
type Choice struct {
	Tokens Tokens
	Icons  Icons
}

// The values a saved preference (the Style section of the Settings tab) may
// take; "auto" and "" both mean "decide from the terminal".
const Auto = "auto"

// ThemeNames and IconSetNames are the choices the Style section offers, "auto" first.
var (
	ThemeNames   = []string{Auto, "dark", "light", "plain"}
	IconSetNames = []string{Auto, "unicode", "ascii"}
)

// Choose picks the theme from the environment: NO_COLOR (any value) gives the
// plain theme; SMITH_THEME=light|dark|plain forces one; otherwise the
// terminal's background decides. SMITH_ASCII=1, or TERM=dumb or linux, gives
// the ASCII symbols. getenv is os.Getenv in production.
func Choose(getenv func(string) string, darkBackground bool) Choice {
	return ChooseWith(getenv, darkBackground, "", "")
}

// ChooseWith is Choose with what the user saved (theme: dark, light or plain;
// icons: unicode or ascii; anything else, "auto" included, means nothing was
// saved). The order is: the environment variables win, then the saved
// choice, then what the terminal says.
func ChooseWith(getenv func(string) string, darkBackground bool, savedTheme, savedIcons string) Choice {
	c := Choice{Tokens: Dark, Icons: Unicode}
	if !darkBackground {
		c.Tokens = Light
	}
	switch strings.ToLower(savedTheme) {
	case "light":
		c.Tokens = Light
	case "dark":
		c.Tokens = Dark
	case "plain":
		c.Tokens = Plain
	}
	if strings.ToLower(savedIcons) == "ascii" {
		c.Icons = ASCII
	}
	switch strings.ToLower(getenv("SMITH_THEME")) {
	case "light":
		c.Tokens = Light
	case "dark":
		c.Tokens = Dark
	case "plain":
		c.Tokens = Plain
	}
	if getenv("NO_COLOR") != "" {
		c.Tokens = Plain
	}
	if term := getenv("TERM"); getenv("SMITH_ASCII") != "" || term == "dumb" || term == "linux" {
		c.Icons = ASCII
	}
	return c
}

// Overrides names the environment variable that decides the theme and the one
// that decides the icons, "" for one that is not set: they win over a saved
// choice, so the Style section says so.
func Overrides(getenv func(string) string) (themeVar, iconsVar string) {
	switch {
	case getenv("NO_COLOR") != "":
		themeVar = "NO_COLOR"
	case slices.Contains(ThemeNames[1:], strings.ToLower(getenv("SMITH_THEME"))):
		themeVar = "SMITH_THEME"
	}
	switch term := getenv("TERM"); {
	case getenv("SMITH_ASCII") != "":
		iconsVar = "SMITH_ASCII"
	case term == "dumb" || term == "linux":
		iconsVar = "TERM=" + term
	}
	return themeVar, iconsVar
}
