package discovery

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

// errFound stops the walk of HasTestFiles at the first test file.
var errFound = errors.New("found")

// HasTestFiles reports whether the project folder holds a file pytest
// collects by its default rule: test_*.py or *_test.py. Hidden folders, the
// venv, build output and sdist staging folders are skipped, as in
// ScanImports.
//
// Parameters: dir is the project folder on disk.
// Returns: true when at least one test file is there; false for a folder that
// cannot be read.
func HasTestFiles(dir string) bool {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if path != dir && (skippedDirs[name] || strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".egg-info") || IsSdistStaging(name)) {
				return filepath.SkipDir
			}
			return nil
		}
		if IsTestFile(name) {
			return errFound
		}
		return nil
	})
	return errors.Is(err, errFound)
}

// IsTestFile reports whether a file name matches pytest's default
// python_files rule: test_*.py or *_test.py.
//
// Parameters: name is the file's base name.
// Returns: true for a test file name.
func IsTestFile(name string) bool {
	stem, ok := strings.CutSuffix(name, ".py")
	return ok && (strings.HasPrefix(stem, "test_") || strings.HasSuffix(stem, "_test"))
}
