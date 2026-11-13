package functionhost

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// assembledFiles are the files that a function host wants next to the compiled
// code. local.settings.json is the Azure Functions Core Tools settings file: it
// is copied only when the person made one, so that `func start` finds it.
var assembledFiles = []string{"package.json", "host.json", "local.settings.json"}

// AssembleFunctionApp copies the manifest files of a function app next to its
// compiled code, in the build folder. That folder is what the function host
// runs and what the zip holds. A file the project does not have is skipped.
//
// Parameters: projectDir is the project folder; buildDir the folder the
// compiler wrote the code into.
// Returns an error when the folder or a copy cannot be written.
func AssembleFunctionApp(projectDir, buildDir string) error {
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return err
	}
	for _, name := range assembledFiles {
		err := copyFile(filepath.Join(projectDir, name), filepath.Join(buildDir, name))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// copyFile copies one file and keeps its mode.
func copyFile(from, to string) error {
	raw, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	info, err := os.Stat(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, raw, info.Mode().Perm())
}
