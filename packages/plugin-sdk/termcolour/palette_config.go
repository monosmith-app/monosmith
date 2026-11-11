package termcolour

// The SGR codes of smith's log palette (the cli-output-design skill): green
// for passed, yellow for a warning or a skip caused by something else, red for
// failed, dim for unchanged, skipped or not reached.
const (
	Green  = "32"
	Yellow = "33"
	Red    = "31"
	Dim    = "2"
)

// Wrap wraps text in one SGR code and a reset. It never changes the text.
//
// Parameters: text is what to colour; code is one of the palette codes.
// Returns: ESC [ code m, text, ESC [ 0 m.
func Wrap(text, code string) string {
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}
