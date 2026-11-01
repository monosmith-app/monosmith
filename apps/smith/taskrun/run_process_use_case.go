package taskrun

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Spec is one command to run.
type Spec struct {
	Argv []string
	Dir  string
	Env  []string
	Out  io.Writer
}

// RunProcess runs a command to completion and returns its exit code. When ctx
// is cancelled the whole process tree is stopped.
func RunProcess(ctx context.Context, spec Spec) (int, error) {
	if len(spec.Argv) == 0 {
		return -1, errors.New("task has no command")
	}
	program := spec.Argv[0]
	if !strings.ContainsAny(program, `/\`) {
		resolved, err := lookPath(program, spec.Env)
		if err != nil {
			return -1, errors.New(MissingProgramMessage(program))
		}
		program = resolved
	} else {
		program = filepath.FromSlash(program)
		if !filepath.IsAbs(program) {
			program = filepath.Join(spec.Dir, program)
		}
	}
	cmd := exec.Command(program, spec.Argv[1:]...)
	cmd.Dir = spec.Dir
	cmd.Env = spec.Env
	cmd.Stdout = spec.Out
	cmd.Stderr = spec.Out
	cmd.Stdin = nil
	prepare(cmd)
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	release, _ := track(cmd)
	defer release()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return exitCode(err)
	case <-ctx.Done():
		killTree(cmd)
		<-done
		return -1, ctx.Err()
	}
}

// lookPath finds a program on the PATH of the child's environment (which
// may put a tool directory first), falling back to smith's own PATH.
func lookPath(program string, env []string) (string, error) {
	var path string
	for _, kv := range env {
		name, value, _ := strings.Cut(kv, "=")
		if name == "PATH" || (isWindows && strings.EqualFold(name, "PATH")) {
			path = value
		}
	}
	if path == "" {
		return exec.LookPath(program)
	}
	exts := []string{""}
	if isWindows && filepath.Ext(program) == "" {
		exts = strings.Split(strings.ToLower(os.Getenv("PATHEXT")), ";")
		if len(exts) == 1 && exts[0] == "" {
			exts = []string{".com", ".exe", ".bat", ".cmd"}
		}
	}
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			continue
		}
		for _, ext := range exts {
			candidate := filepath.Join(dir, program+ext)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() && (isWindows || info.Mode()&0o111 != 0) {
				return candidate, nil
			}
		}
	}
	return exec.LookPath(program)
}

func exitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	return -1, err
}

// Environment builds a child environment: the caller's, plus extra entries,
// replacing keys case-insensitively on Windows where Path and PATH are one.
func Environment(extra map[string]string) []string {
	env := os.Environ()
	for k, v := range extra {
		replaced := false
		for i, kv := range env {
			name, _, _ := strings.Cut(kv, "=")
			if name == k || (isWindows && strings.EqualFold(name, k)) {
				env[i] = name + "=" + v
				replaced = true
			}
		}
		if !replaced {
			env = append(env, k+"="+v)
		}
	}
	return env
}
