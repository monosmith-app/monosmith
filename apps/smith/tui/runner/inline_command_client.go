package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"golang.org/x/term"
)

// Inline is a command that runs in the real terminal, the way it would from a
// shell: it inherits the terminal, so its output and live status line go to the
// scrollback, and they stay there when the interface ends. It implements
// Bubble Tea's ExecCommand, which hands the terminal over for the run and takes
// it back after.
type Inline struct {
	Argv []string

	run   func(stdin io.Reader, stdout, stderr io.Writer) (int, error)
	pause func(stdin io.Reader)
	now   func() time.Time

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	// Result is set once Run returned.
	Result InlineResult
}

// InlineResult is how an inline run ended.
type InlineResult struct {
	Exit Exit
	Took time.Duration
}

// NewInline prepares to run smith's arguments in the directory of the spec
// with the terminal of this process (Spec.Program empty means this very smith).
func NewInline(spec Spec) *Inline {
	i := &Inline{Argv: spec.Argv, pause: waitForKey, now: time.Now}
	i.run = func(stdin io.Reader, stdout, stderr io.Writer) (int, error) {
		program := spec.Program
		if program == "" {
			self, err := os.Executable()
			if err != nil {
				return -1, err
			}
			program = self
		}
		cmd := exec.Command(program, spec.Argv...)
		cmd.Dir = spec.Dir
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
		err := cmd.Run()
		var ee *exec.ExitError
		switch {
		case err == nil:
			return 0, nil
		case errors.As(err, &ee):
			return ee.ExitCode(), nil
		}
		return -1, err
	}
	return i
}

// NewInlineWith replaces the process and the wait for a key, so tests need
// neither a terminal nor a child.
func NewInlineWith(argv []string, run func(stdin io.Reader, stdout, stderr io.Writer) (int, error), pause func()) *Inline {
	return &Inline{
		Argv: argv, now: time.Now, run: run,
		pause: func(io.Reader) {
			if pause != nil {
				pause()
			}
		},
	}
}

// WithClock replaces the clock the duration is measured with.
func (i *Inline) WithClock(now func() time.Time) *Inline { i.now = now; return i }

func (i *Inline) SetStdin(r io.Reader)  { i.stdin = r }
func (i *Inline) SetStdout(w io.Writer) { i.stdout = w }
func (i *Inline) SetStderr(w io.Writer) { i.stderr = w }

// Run prints the command, runs it, prints how it ended and waits for a key, so
// the output can be read before the interface comes back. It never fails: the
// outcome is in Result.
func (i *Inline) Run() error {
	out, errw := i.stdout, i.stderr
	if out == nil {
		out = io.Discard
	}
	if errw == nil {
		errw = io.Discard
	}
	fmt.Fprintf(out, "\r\n$ %s\r\n", commands.Line(i.Argv))

	// Ctrl+C reaches the child and this process alike; the child stops, and this
	// process stays to report it.
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)
	defer signal.Stop(interrupted)

	start := i.now()
	code, err := i.run(i.stdin, out, errw)
	exit := Exit{Code: code, Err: err}
	select {
	case <-interrupted:
		exit.Cancelled = true
	default:
	}
	i.Result = InlineResult{Exit: exit, Took: i.now().Sub(start)}

	fmt.Fprintf(out, "\r\n%s\r\n", InlineSummary(i.Argv, i.Result))
	fmt.Fprint(out, "Press any key to return to smith…")
	if i.pause != nil {
		i.pause(i.stdin)
	}
	fmt.Fprint(out, "\r\n")
	return nil
}

// InlineSummary is the one-line outcome with the command that reproduces it.
func InlineSummary(argv []string, r InlineResult) string {
	took := r.Took.Round(100 * time.Millisecond)
	line := commands.Line(argv)
	switch {
	case r.Exit.Cancelled:
		return fmt.Sprintf("■ stopped after %s · %s", took, line)
	case r.Exit.Err != nil:
		return fmt.Sprintf("✖ could not run: %v · %s", r.Exit.Err, line)
	case r.Exit.Code != 0:
		return fmt.Sprintf("✖ failed (exit %d) in %s · %s", r.Exit.Code, took, line)
	}
	return fmt.Sprintf("✔ done in %s · %s", took, line)
}

// waitForKey reads one key press, putting a terminal in raw mode for it.
func waitForKey(stdin io.Reader) {
	f, ok := stdin.(*os.File)
	if !ok {
		f = os.Stdin
	}
	if fd := int(f.Fd()); term.IsTerminal(fd) {
		if old, err := term.MakeRaw(fd); err == nil {
			defer func() { _ = term.Restore(fd, old) }()
		}
	}
	buf := make([]byte, 1)
	_, _ = f.Read(buf)
}
