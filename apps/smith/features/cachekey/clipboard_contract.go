package cachekey

// Clipboard is the user's clipboard, as far as smith can reach it.
type Clipboard interface {
	// Name says how the clipboard is reached ("clip", "pbcopy", "the terminal").
	Name() string
	// Write replaces the clipboard's text; "" clears it.
	Write(text string) error
	// Read returns the clipboard's text. CanRead tells whether it works.
	Read() (string, error)
	// CanRead is false for a clipboard that can only be written, such as the
	// terminal's: a copy there cannot be checked, so it is never cleared.
	CanRead() bool
}
