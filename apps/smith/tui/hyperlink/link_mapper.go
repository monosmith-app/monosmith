package hyperlink

import "strings"

// Link draws text as a link to url when links are on; otherwise it returns the
// text alone. The sequence ends with ESC \ (ST), which the layout code of the
// shell keeps. Control characters in the address are dropped, so a name cannot
// end the sequence early.
//
// Parameters: url is the address; text is what shows; on is Supported.
// Returns: the text to draw.
func Link(url, text string, on bool) string {
	url = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, url)
	if !on || url == "" {
		return text
	}
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}
