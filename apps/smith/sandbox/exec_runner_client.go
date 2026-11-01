package sandbox

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
)

// ExecRunner runs this smith binary with args in dir. It is the Runner
// `smith try` uses; a test must not call it (the binary would start itself).
func ExecRunner(dir string, env []string, args ...string) (string, int, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", 0, err
	}
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	err = cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return buf.String(), exit.ExitCode(), nil
	}
	return buf.String(), 0, err
}
