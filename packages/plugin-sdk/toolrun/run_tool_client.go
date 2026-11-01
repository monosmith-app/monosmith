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
)

// Log is where tool output goes; tests may replace it.
var Log io.Writer = os.Stderr

// Run executes argv in dir, streaming its output to Log. extraEnv entries are
// KEY=VALUE pairs added to the plugin's environment.
func Run(dir string, argv []string, extraEnv ...string) error {
	cmd, err := command(dir, argv, extraEnv)
	if err != nil {
		return err
	}
	cmd.Stdout = Log
	cmd.Stderr = Log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return nil
}

// Output executes argv in dir and returns its stdout; stderr goes to Log.
func Output(dir string, argv []string, extraEnv ...string) (string, error) {
	cmd, err := command(dir, argv, extraEnv)
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
// caller that tells failures apart by what the tool printed.
func Capture(dir string, argv []string, extraEnv ...string) (stdout, stderr string, err error) {
	cmd, err := command(dir, argv, extraEnv)
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
	cmd.Env = append(os.Environ(), extraEnv...)
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

func command(dir string, argv []string, extraEnv []string) (*exec.Cmd, error) {
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return nil, fmt.Errorf("%s is not installed or not on PATH", argv[0])
	}
	cmd := exec.Command(path, argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd, nil
}
