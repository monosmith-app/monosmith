package consolecolour

import (
	"os"

	"golang.org/x/sys/windows"
)

// Enable turns on virtual terminal processing for stdout and stderr when they
// are a console, so colour codes render instead of showing as text. A handle
// that is not a console (a pipe, a file) is left alone, and a console that
// refuses the mode (before Windows 10) keeps printing as before.
func Enable() {
	for _, f := range []*os.File{os.Stdout, os.Stderr} {
		h := windows.Handle(f.Fd())
		var mode uint32
		if windows.GetConsoleMode(h, &mode) != nil {
			continue
		}
		_ = windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}
