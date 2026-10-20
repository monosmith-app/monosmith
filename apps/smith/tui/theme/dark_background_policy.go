package theme

import (
	"strconv"
	"strings"
)

// DarkBackground guesses whether the terminal has a dark background from
// COLORFGBG ("15;0": foreground;background as ANSI colour numbers), which
// many terminals set. It does not query the terminal, which can stall; with no
// hint it assumes dark. SMITH_THEME=light|dark overrides it (see Choose).
func DarkBackground(getenv func(string) string) bool {
	parts := strings.Split(getenv("COLORFGBG"), ";")
	if len(parts) < 2 {
		return true
	}
	bg, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return true
	}
	return bg != 7 && (bg < 9 || bg == 8)
}
