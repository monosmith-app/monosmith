package cachecommand

import (
	"io"
	"os"

	"golang.org/x/term"
)

// stdinFeed is the one reader of standard input for a command that mixes
// typed lines with single key presses. The terminal is in raw mode for the
// whole command, so one goroutine owns the input and no key press is lost
// between a line and a one-key choice.
type stdinFeed struct{ ch chan rune }

// newStdinFeed starts reading r, a byte at a time, until it ends.
func newStdinFeed(r io.Reader) *stdinFeed {
	f := &stdinFeed{ch: make(chan rune)}
	go func() {
		defer close(f.ch)
		buf := make([]byte, 1)
		for {
			if n, err := r.Read(buf); err != nil || n == 0 {
				return
			}
			f.ch <- rune(buf[0])
		}
	}()
	return f
}

// Keys delivers the key presses; it closes when the input ends.
func (f *stdinFeed) Keys() <-chan rune { return f.ch }

// rawTerminal puts the terminal in raw mode, where each key press arrives at
// once and nothing is echoed. restore puts it back; ok is false when the
// terminal could not be switched.
func rawTerminal() (restore func(), ok bool) {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}, false
	}
	return func() { _ = term.Restore(fd, state) }, true
}

// lineEditor reads typed lines from a feed in raw mode: it echoes the
// characters, handles backspace, and ends a line at Enter.
type lineEditor struct {
	feed *stdinFeed
	echo io.Writer
	// interrupt is called on Ctrl-C.
	interrupt func()
	pending   []byte
}

// Read returns one typed line, with a newline, per call.
func (e *lineEditor) Read(p []byte) (int, error) {
	if len(e.pending) == 0 {
		var line []rune
		for {
			r, ok := <-e.feed.ch
			if !ok {
				return 0, io.EOF
			}
			switch {
			case r == '\r' || r == '\n':
				_, _ = io.WriteString(e.echo, "\n")
				e.pending = append([]byte(string(line)), '\n')
			case r == 3:
				e.interrupt()
				continue
			case r == 0x7f || r == 8:
				if len(line) > 0 {
					line = line[:len(line)-1]
					_, _ = io.WriteString(e.echo, "\b \b")
				}
				continue
			case r < 32:
				continue
			default:
				line = append(line, r)
				_, _ = io.WriteString(e.echo, string(r))
				continue
			}
			break
		}
	}
	n := copy(p, e.pending)
	e.pending = e.pending[n:]
	return n, nil
}
