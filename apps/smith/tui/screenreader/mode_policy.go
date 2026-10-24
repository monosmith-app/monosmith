package screenreader

import "strings"

// Variable is the environment variable that asks for this mode.
const Variable = "SMITH_SCREEN_READER"

// ThemeVariable and PlainTheme name the setting that asks for no colour and
// no decoration: SMITH_THEME=plain is also a request for plain output.
const (
	ThemeVariable = "SMITH_THEME"
	PlainTheme    = "plain"
)

// Wanted reports whether the environment asks for the screen-reader mode:
// SMITH_SCREEN_READER set to anything but empty, 0, false or no.
func Wanted(getenv func(string) string) bool {
	switch strings.ToLower(strings.TrimSpace(getenv(Variable))) {
	case "", "0", "false", "no", "off":
		return false
	}
	return true
}

// Plain reports whether a bare smith should ask its questions as plain
// numbered lines instead of drawing the full-screen shell: when
// SMITH_SCREEN_READER asks for it, when SMITH_THEME=plain, or when standard
// output is not a terminal (the shell could not be drawn there).
func Plain(getenv func(string) string, stdoutIsTerminal bool) bool {
	if Wanted(getenv) || strings.EqualFold(strings.TrimSpace(getenv(ThemeVariable)), PlainTheme) {
		return true
	}
	return !stdoutIsTerminal
}
