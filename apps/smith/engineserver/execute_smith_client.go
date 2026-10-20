package engineserver

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
)

// Execute runs smith with the arguments in dir and returns what it wrote to
// stdout and its exit code. When onLine is set, each line of either stream is
// also passed to it as it is written.
type Execute func(ctx context.Context, dir string, args []string, onLine func(stream runevents.Stream, line string)) (stdout []byte, exitCode int, err error)

// ExecuteSelf runs the smith that is running, as `smith mcp` does: not a
// terminal and in CI mode, so no command stops to ask.
func ExecuteSelf(ctx context.Context, dir string, args []string, onLine func(runevents.Stream, string)) ([]byte, int, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, -1, err
	}
	cmd := exec.CommandContext(ctx, self, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CI=1", "NO_COLOR=1")
	var stdout, stderr bytes.Buffer
	var mu sync.Mutex
	feed := func(stream runevents.Stream, capture *bytes.Buffer) io.Writer {
		return &lineWriter{capture: capture, mu: &mu, line: func(l string) {
			if onLine != nil {
				onLine(stream, l)
			}
		}}
	}
	so, se := feed(runevents.Stdout, &stdout), feed(runevents.Stderr, &stderr)
	cmd.Stdout, cmd.Stderr = so, se
	runErr := cmd.Run()
	so.(*lineWriter).flush()
	se.(*lineWriter).flush()
	if runErr != nil {
		var exit *exec.ExitError
		if errors.As(runErr, &exit) && ctx.Err() == nil {
			return stdout.Bytes(), exit.ExitCode(), nil
		}
		return stdout.Bytes(), -1, runErr
	}
	return stdout.Bytes(), 0, nil
}

// lineWriter keeps everything written and reports it a line at a time.
type lineWriter struct {
	capture *bytes.Buffer
	mu      *sync.Mutex
	line    func(string)
	pending bytes.Buffer
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.capture.Write(p)
	w.pending.Write(p)
	for {
		i := bytes.IndexByte(w.pending.Bytes(), '\n')
		if i < 0 {
			return len(p), nil
		}
		raw, _ := w.pending.ReadBytes('\n')
		w.line(string(bytes.TrimRight(raw, "\r\n")))
	}
}

func (w *lineWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pending.Len() > 0 {
		sc := bufio.NewScanner(&w.pending)
		for sc.Scan() {
			w.line(sc.Text())
		}
	}
}
