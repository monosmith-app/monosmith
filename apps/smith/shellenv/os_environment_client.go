package shellenv

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// OSEnvironment is the Environment of the machine smith runs on.
type OSEnvironment struct{}

// Getenv reads an environment variable.
func (OSEnvironment) Getenv(key string) string { return os.Getenv(key) }

// OS is runtime.GOOS.
func (OSEnvironment) OS() string { return runtime.GOOS }

// HomeDir is the home folder of the user, "" when it is not known.
func (OSEnvironment) HomeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// ReadFile reads a file.
func (OSEnvironment) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

// WriteFile writes a file, with the folders above it, for the user only to
// change.
func (OSEnvironment) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Remove deletes a file; a file that is not there is not an error.
func (OSEnvironment) Remove(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// Exists is true when something is at the path.
func (OSEnvironment) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ParentProgram is the name of the program that started smith.
func (OSEnvironment) ParentProgram() string { return parentProgramName(os.Getppid()) }
