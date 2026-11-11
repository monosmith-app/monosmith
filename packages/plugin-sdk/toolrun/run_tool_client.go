package toolrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// Log is where tool output goes; tests may replace it.
var Log io.Writer = os.Stderr

// Run executes argv in dir, streaming its output to Log. extraEnv entries are
// KEY=VALUE pairs added to the plugin's environment. When the host handed on
// tool colour (ToolColour), the tool is asked for colour, only its SGR codes
// pass and smith paint applies; otherwise it is asked for plain text and
// every escape code is removed.
func Run(dir string, argv []string, extraEnv ...string) error {
	colour := ToolColour()
	cmd, err := command(dir, argv, extraEnv, colour)
	if err != nil {
		return err
	}
	out := termcolour.NewLineWriter(Log, colour)
	defer out.Flush()
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return nil
}

// Output executes argv in dir and returns its stdout; stderr goes to Log. The
// tool is asked for plain text: its output is read, not shown.
func Output(dir string, argv []string, extraEnv ...string) (string, error) {
	cmd, err := command(dir, argv, extraEnv, false)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = Log
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return out.String(), nil
}

// Capture executes argv in dir and returns its stdout and stderr, for a
// caller that tells failures apart by what the tool printed. The tool is
// asked for plain text.
func Capture(dir string, argv []string, extraEnv ...string) (stdout, stderr string, err error) {
	cmd, err := command(dir, argv, extraEnv, false)
	if err != nil {
		return "", "", err
	}
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), errOut.String(), fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return out.String(), errOut.String(), nil
}

// ErrTimeout is wrapped by the error CaptureTimeout returns when the tool did
// not finish in time and was stopped.
var ErrTimeout = errors.New("timed out")

// CaptureTimeout is Capture with a limit: a tool still running after timeout is
// killed and the error wraps ErrTimeout. Only the tool is stopped.
func CaptureTimeout(dir string, argv []string, timeout time.Duration, extraEnv ...string) (stdout, stderr string, err error) {
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return "", "", fmt.Errorf("%s is not installed or not on PATH", argv[0])
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, argv[1:]...)
	cmd.Dir = dir
	cmd.Env = toolEnv(extraEnv, false)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	// A killed npm can leave a child holding the pipes open; do not wait for it.
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return out.String(), errOut.String(), fmt.Errorf("%s %w after %s", strings.Join(argv, " "), ErrTimeout, timeout)
		}
		return out.String(), errOut.String(), fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return out.String(), errOut.String(), nil
}

// Available reports whether a program can be found on PATH (honouring
// PATHEXT on Windows, so "npm" finds npm.cmd).
func Available(program string) bool {
	_, err := exec.LookPath(program)
	return err == nil
}

func command(dir string, argv []string, extraEnv []string, colour bool) (*exec.Cmd, error) {
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return nil, fmt.Errorf("%s is not installed or not on PATH", argv[0])
	}
	cmd := exec.Command(path, argv[1:]...)
	cmd.Dir = dir
	cmd.Env = toolEnv(extraEnv, colour)
	return cmd, nil
}

// toolEnv is the plugin's environment, then the entries that ask the tool
// for colour or plain text, then the caller's: os/exec keeps the last value of
// a repeated key.
func toolEnv(extraEnv []string, colour bool) []string {
	env := append(os.Environ(), termcolour.Env(colour)...)
	return append(env, extraEnv...)
}

// ToolColour reports whether the tools a plugin runs stream to a person who
// sees colour. A plugin's stdout is the protocol pipe, so it relies on what
// the host handed on in termcolour.ToolColourEnv.
//
// Returns: true when Run asks tools for colour.
func ToolColour() bool { return termcolour.ToolWanted(os.Getenv, false) }

// Probe executes argv in dir to learn whether a tool exists and what it
// prints, for a check such as `cargo audit --version`. It returns stdout and
// discards stderr: a missing tool is the check's finding to word, and the
// tool's own "no such command" lines would reach the person before it.
//
// Parameters: dir is where the tool runs; argv is the program and arguments;
// extraEnv adds environment entries.
// Returns: stdout, and an error when the tool is absent or exits non-zero.
func Probe(dir string, argv []string, extraEnv ...string) (string, error) {
	out, _, err := Capture(dir, argv, extraEnv...)
	return out, err
}
