package hyperlink

import (
	"strconv"
	"strings"
)

// hyperlinkPrograms are the TERM_PROGRAM values of terminals that draw OSC 8
// hyperlinks.
var hyperlinkPrograms = map[string]bool{"iTerm.app": true, "vscode": true, "WezTerm": true, "ghostty": true, "Hyper": true}

// Supported reports whether the terminal is known to draw OSC 8 hyperlinks.
// The answer comes from the environment only, never from a query to the
// terminal (a terminal that ignores a query stalls the screen). FORCE_HYPERLINK
// set to 1 turns links on and to 0 turns them off. A terminal that is not
// known, and a multiplexer (tmux, screen), get the plain address.
//
// Parameters: getenv is os.Getenv in production.
// Returns: true when links can be drawn.
func Supported(getenv func(string) string) bool {
	switch getenv("FORCE_HYPERLINK") {
	case "0":
		return false
	case "1":
		return true
	}
	term := getenv("TERM")
	if term == "dumb" || getenv("TMUX") != "" || getenv("STY") != "" {
		return false
	}
	switch {
	case getenv("WT_SESSION") != "", getenv("KITTY_WINDOW_ID") != "", getenv("KONSOLE_VERSION") != "", getenv("WEZTERM_EXECUTABLE") != "":
		return true
	case hyperlinkPrograms[getenv("TERM_PROGRAM")]:
		return true
	case strings.Contains(term, "kitty"), strings.Contains(term, "alacritty"), strings.HasPrefix(term, "foot"):
		return true
	}
	vte, err := strconv.Atoi(getenv("VTE_VERSION"))
	return err == nil && vte >= 5000
}
