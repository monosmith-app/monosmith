package clipboard

import (
	"encoding/base64"
	"errors"
	"io"
)

// maxBytes is the most text sent in one sequence; terminals drop longer ones.
const maxBytes = 74000

// ErrNothing is returned when there is no text to copy or nowhere to write it.
var ErrNothing = errors.New("nothing to copy")

// Copy asks the terminal behind w to put text on the clipboard. The whole
// sequence goes out in one Write so it cannot be split by other output.
func Copy(w io.Writer, text string) error {
	if w == nil || text == "" {
		return ErrNothing
	}
	if len(text) > maxBytes {
		text = text[:maxBytes]
	}
	_, err := io.WriteString(w, "\x1b]52;c;"+base64.StdEncoding.EncodeToString([]byte(text))+"\x07")
	return err
}
