package shellenv

import (
	"os"
	"path/filepath"
	"testing"
)

// tempEnv is an Environment whose home is a temporary folder. The files are
// real, so the tests see what install writes; the variables, the operating
// system and the parent program are what the test sets.
type tempEnv struct {
	OSEnvironment
	home   string
	vars   map[string]string
	goos   string
	parent string
}

// newTempEnv makes an environment for a shell on an operating system.
func newTempEnv(t *testing.T, goos string) *tempEnv {
	t.Helper()
	return &tempEnv{home: t.TempDir(), vars: map[string]string{}, goos: goos}
}

func (e *tempEnv) Getenv(key string) string { return e.vars[key] }
func (e *tempEnv) OS() string               { return e.goos }
func (e *tempEnv) HomeDir() string          { return e.home }
func (e *tempEnv) ParentProgram() string    { return e.parent }

// put writes a file below the home folder and returns its path.
func (e *tempEnv) put(t *testing.T, content string, parts ...string) string {
	t.Helper()
	p := filepath.Join(append([]string{e.home}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// read returns a file, or "" when it is not there.
func (e *tempEnv) read(path string) string {
	raw, _ := os.ReadFile(path)
	return string(raw)
}
