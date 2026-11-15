package statuspaint

import "io"

// writer paints each write it gets before passing it on.
type writer struct{ out io.Writer }

// Write paints the lines of p and writes them. It reports len(p), as a writer
// must, though the painted text is longer.
func (w writer) Write(p []byte) (int, error) {
	if _, err := io.WriteString(w.out, Paint(string(p))); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Writer paints the status lines a command writes to out, when colour is wanted
// there (see WantedHere); otherwise it returns out itself, so a caller that
// checks whether out is a terminal still sees the file. A command writes its
// lines whole (fmt.Fprintf with a newline), which is what the painting needs.
//
// Parameters: out is where the command writes.
// Returns: a writer that paints, or out.
func Writer(out io.Writer) io.Writer {
	if !WantedHere(out) {
		return out
	}
	return writer{out: out}
}
