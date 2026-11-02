package runner

import (
	"bufio"
	"context"
	"errors"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Spec is the command to start: smith's arguments (without the program name)
// and the directory to run in.
type Spec struct {
	Argv []string
	Dir  string
	// Program is the executable; empty means this very smith (os.Executable).
	Program string
}

// CommandLine is the line as a user would type it, for display.
func (s Spec) CommandLine() string { return commands.Line(s.Argv) }

// Exit is how a process ended.
type Exit struct {
	Code      int
	Cancelled bool
	Err       error
}

// Handle is what the shell needs of a running command, so tests can replace
// the process with a script.
type Handle interface {
	// Next blocks for output and, at the end, the exit (see Process.Next).
	Next() (lines []string, exit *Exit)
	// Stop asks the command to stop.
	Stop()
}

// Process is a running child. Lines arrive through Next.
type Process struct {
	lines  chan string
	done   chan Exit
	cancel context.CancelFunc
	once   sync.Once
	// stopped is set by Stop so the exit can be reported as cancelled.
	mu      sync.Mutex
	stopped bool
}

// maxBatch bounds how many lines one Next returns, so the screen redraws
// between bursts of output.
const maxBatch = 200

// Start launches the command and returns at once; the process runs on its own
// goroutines.
func Start(parent context.Context, spec Spec) (*Process, error) {
	program := spec.Program
	if program == "" {
		self, err := os.Executable()
		if err != nil {
			return nil, err
		}
		program = self
	}
	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, program, spec.Argv...)
	cmd.Dir = spec.Dir
	// The child is not a terminal: it prints plain lines, one per event.
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "SMITH_INTERFACE=tui")
	cmd.Stdin = nil
	// Ask the child to stop the way a terminal would (Ctrl+C), so it can clean
	// up its own task processes; where that is not possible, kill it.
	cmd.Cancel = func() error {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	cmd.WaitDelay = 5 * time.Second
	pr, pw := io.Pipe()
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	p := &Process{lines: make(chan string, 1024), done: make(chan Exit, 1), cancel: cancel}
	go func() {
		sc := bufio.NewScanner(pr)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)
		for sc.Scan() {
			p.lines <- sc.Text()
		}
		close(p.lines)
	}()
	go func() {
		err := cmd.Wait()
		_ = pw.Close()
		p.mu.Lock()
		stopped := p.stopped
		p.mu.Unlock()
		exit := Exit{Cancelled: stopped}
		var ee *exec.ExitError
		switch {
		case err == nil:
		case errors.As(err, &ee):
			exit.Code = ee.ExitCode()
		default:
			exit.Code, exit.Err = -1, err
		}
		p.done <- exit
	}()
	return p, nil
}

// Next blocks until the process printed something or ended. It returns the
// lines available (at most a batch) and, once the output is drained and the
// process has exited, the exit. Call it again until exit is not nil.
func (p *Process) Next() (lines []string, exit *Exit) {
	line, ok := <-p.lines
	if !ok {
		e := <-p.done
		return nil, &e
	}
	lines = append(lines, line)
	for len(lines) < maxBatch {
		select {
		case l, ok := <-p.lines:
			if !ok {
				return lines, nil
			}
			lines = append(lines, l)
		default:
			return lines, nil
		}
	}
	return lines, nil
}

// Stop asks the process to stop: Ctrl+C where the platform allows, a kill after
// a grace period or where it does not.
func (p *Process) Stop() {
	p.once.Do(func() {
		p.mu.Lock()
		p.stopped = true
		p.mu.Unlock()
		p.cancel()
	})
}
