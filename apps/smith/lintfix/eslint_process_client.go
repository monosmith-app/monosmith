package lintfix

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// runProcess is the Runner that starts the program: its two outputs apart, so
// ESLint's JSON report is not mixed with a warning it prints.
func runProcess(dir string, argv []string) (string, string, int, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return stdout.String(), stderr.String(), exit.ExitCode(), nil
	}
	return stdout.String(), stderr.String(), 0, err
}

// installedESLint is the version of the ESLint under root's node_modules, ""
// when none is installed.
func installedESLint(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, "node_modules", "eslint", "package.json"))
	if err != nil {
		return ""
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(raw, &manifest) != nil {
		return ""
	}
	return manifest.Version
}
