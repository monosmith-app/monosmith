package cachekey

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// tempEnv is an Environment whose home is a temporary folder. The files are
// real; the variables, the operating system and the parent program are what
// the test sets.
type tempEnv struct {
	shellenv.OSEnvironment
	home   string
	vars   map[string]string
	goos   string
	parent string
}

// newTempEnv makes an environment for a shell on an operating system.
func newTempEnv(t *testing.T, goos, parent string) *tempEnv {
	t.Helper()
	return &tempEnv{home: t.TempDir(), vars: map[string]string{}, goos: goos, parent: parent}
}

func (e *tempEnv) Getenv(key string) string { return e.vars[key] }
func (e *tempEnv) OS() string               { return e.goos }
func (e *tempEnv) HomeDir() string          { return e.home }
func (e *tempEnv) ParentProgram() string    { return e.parent }

// put writes a file below the home folder with a mode and returns its path.
func (e *tempEnv) put(t *testing.T, content string, mode os.FileMode, parts ...string) string {
	t.Helper()
	p := filepath.Join(append([]string{e.home}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	return p
}

// read returns a file, or "" when it is not there.
func (e *tempEnv) read(path string) string {
	raw, _ := os.ReadFile(path)
	return string(raw)
}
