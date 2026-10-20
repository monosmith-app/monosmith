package theme

import "strings"

// Choice says which tokens and icons to use.
type Choice struct {
	Tokens Tokens
	Icons  Icons
}

// Choose picks the theme from the environment: NO_COLOR (any value) gives the
// plain theme; SMITH_THEME=light|dark|plain forces one; otherwise the
// terminal's background decides. SMITH_ASCII=1, or TERM=dumb or linux, gives
// the ASCII symbols. getenv is os.Getenv in production.
func Choose(getenv func(string) string, darkBackground bool) Choice {
	c := Choice{Tokens: Dark, Icons: Unicode}
	if !darkBackground {
		c.Tokens = Light
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
