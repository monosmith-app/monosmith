package plugintrust

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Measure resolves a plugin's command to the files it runs and hashes them:
// the program (found the way the operating system would find it), and every
// later argument that names an existing file, which is how an interpreter's
// script (`node plugin.mjs`) is covered. The command must name a program that
// exists.
func Measure(command []string) (Fingerprint, error) {
	if len(command) == 0 || strings.TrimSpace(command[0]) == "" {
		return Fingerprint{}, fmt.Errorf("the plugin's command is empty")
	}
	program, err := exec.LookPath(command[0])
	if err != nil {
		// A file the operating system will not run by that name (no .exe,
		// no execute bit) is still the file the command names.
		if info, statErr := os.Stat(command[0]); statErr != nil || !info.Mode().IsRegular() {
			return Fingerprint{}, fmt.Errorf("the program %q was not found: %w", command[0], err)
		}
		program = command[0]
	}
	fp := Fingerprint{Command: append([]string(nil), command...)}
	first, err := measureFile(program)
	if err != nil {
		return Fingerprint{}, err
	}
	fp.Executables = append(fp.Executables, first)
	for _, arg := range command[1:] {
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		info, err := os.Stat(arg)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		e, err := measureFile(arg)
		if err != nil {
			return Fingerprint{}, err
		}
		fp.Executables = append(fp.Executables, e)
	}
	return fp, nil
}

func measureFile(path string) (Executable, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Executable{}, err
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		abs = real
	}
	f, err := os.Open(abs)
	if err != nil {
		return Executable{}, err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return Executable{}, fmt.Errorf("reading %s: %w", abs, err)
	}
	return Executable{Path: abs, SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}
