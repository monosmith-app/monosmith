package screenreader

import "strings"

// Variable is the environment variable that asks for this mode.
const Variable = "SMITH_SCREEN_READER"

// Wanted reports whether the environment asks for the screen-reader mode:
// SMITH_SCREEN_READER set to anything but empty, 0, false or no.
func Wanted(getenv func(string) string) bool {
	switch strings.ToLower(strings.TrimSpace(getenv(Variable))) {
	case "", "0", "false", "no", "off":
		return false
	}
	return true
}
