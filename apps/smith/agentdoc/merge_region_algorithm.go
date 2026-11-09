package agentdoc

import (
	"errors"
	"strings"
)

// ErrBrokenMarkers is returned when a file has one marker of the region, or
// the markers in the wrong order: smith cannot tell which lines it owns.
var ErrBrokenMarkers = errors.New("the smith markers of the file are incomplete or in the wrong order")

// Merge puts the generated region into the text of a file.
//
//   - A file that does not exist (exists false) becomes header, a blank line and the region.
//   - A file with both markers keeps every line outside them and gets the region in place.
//   - A file with no marker keeps its text and gets the region at the end.
//
// A file that uses CRLF line ends keeps them. The result ends with a newline.
// Merge returns ErrBrokenMarkers when the file has only one marker, or the end
// before the begin.
func Merge(existing string, exists bool, header, region string) (string, error) {
	if !exists {
		return header + "\n" + region, nil
	}
	crlf := strings.Contains(existing, "\r\n")
	text := strings.ReplaceAll(existing, "\r\n", "\n")
	begin, end := strings.Index(text, BeginMarker), strings.Index(text, EndMarker)
	var out string
	switch {
	case begin < 0 && end < 0:
		out = strings.TrimRight(text, "\n")
		if out != "" {
			out += "\n\n"
		}
		out += region
	case begin < 0 || end < begin:
		return "", ErrBrokenMarkers
	default:
		tail := text[end+len(EndMarker):]
		tail = strings.TrimPrefix(tail, "\n")
		out = text[:begin] + region + tail
	}
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, nil
}
