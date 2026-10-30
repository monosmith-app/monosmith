package declaredimports

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// goCommand runs the go command in dir and returns what it printed and its
// exit code. It is a variable so tests can stand in for the toolchain.
var goCommand = func(dir string, env []string, args ...string) (stdout, stderr string, code int, err error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err = cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return out.String(), errOut.String(), exit.ExitCode(), nil
	}
	return out.String(), errOut.String(), 0, err
}

// Check says whether the go.mod of the module at moduleRoot matches its
// imports. It asks `go mod tidy -diff` (which changes nothing) when the go
// command has it and the module does not lean on go.work, which tidy cannot
// see; otherwise, or when tidy cannot run (no network for a module it needs),
// it compares go.mod with the sources (Compare).
func Check(workspaceRoot, moduleRoot string) (Result, error) {
	result, leans, err := Compare(workspaceRoot, moduleRoot)
	if err != nil {
		return result, err
	}
	if leans {
		result.Method = "go.mod compared with the imports (tidy cannot see go.work, which provides this module's workspace requirements)"
		return result, nil
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(moduleRoot))
	version, _, _, err := goCommand(dir, nil, "version")
	if err != nil || !supportsTidyDiff(version) {
		result.Method = "go.mod compared with the imports (this Go has no go mod tidy -diff)"
		return result, nil
	}
	diff, err := tidyDiff(dir)
	if err != nil {
		result.Method = "go.mod compared with the imports (go mod tidy -diff could not run: " + firstLine(err.Error()) + ")"
		return result, nil
	}
	return Result{Method: "go mod tidy -diff", Diff: diff}, nil
}

// tidyDiff is what `go mod tidy` would change in dir, as a unified diff, or
// "" when go.mod and go.sum are tidy. Neither file is written. The workspace
// is switched off: tidy works on the one module.
func tidyDiff(dir string) (string, error) {
	stdout, stderr, code, err := goCommand(dir, []string{"GOWORK=off"}, "mod", "tidy", "-diff")
	if err != nil {
		return "", err
	}
	switch {
	case code == 0:
		return "", nil
	case code == 1 && strings.TrimSpace(stdout) != "":
		return stdout, nil
	}
	return "", fmt.Errorf("%s", strings.TrimSpace(stderr))
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return line
}
