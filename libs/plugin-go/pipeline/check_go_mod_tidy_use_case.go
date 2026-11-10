package pipeline

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// tidyTimeout limits one `go mod tidy -diff`.
const tidyTimeout = 60 * time.Second

// tidyDiffLines caps how many lines of the diff a finding shows.
const tidyDiffLines = 25

// tidyParallel caps how many modules are checked at once.
const tidyParallel = 4

// runTidy runs `go mod tidy -diff` in dir from the module cache only, so doctor
// never downloads. Tests replace it.
var runTidy = func(dir string) (stdout, stderr string, err error) {
	return toolrun.CaptureTimeout(dir, []string{"go", "mod", "tidy", "-diff"}, tidyTimeout, "GOPROXY=off")
}

// CheckModuleTidy reports whether the go.mod and go.sum of project p are what
// `go mod tidy` would write (a direct import marked `// indirect`, or a needed
// requirement missing, makes gopls call the module out of sync).
//
// Parameters: workspaceRoot is the workspace folder; p is the Go project.
// Returns: the finding, and false when the check does not apply (no go.mod, or
// go is not on PATH, which the tool findings report). A diff is a warning with
// a fix. When the check cannot run (Go older than 1.23, modules missing from
// the cache, a timeout), the finding passes and its Detail says why.
func CheckModuleTidy(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		return protocol.Finding{}, false
	}
	if _, err := exec.LookPath("go"); err != nil {
		return protocol.Finding{}, false
	}
	f := protocol.Finding{Check: p.Root + " go.mod and go.sum are tidy", Project: p.Name, Remedy: "run `go mod tidy` in " + p.Root}
	stdout, stderr, err := runTidy(dir)
	switch {
	case err == nil:
		f.OK = true
	case strings.TrimSpace(stdout) != "" && !errors.Is(err, toolrun.ErrTimeout):
		f.Warning = true
		f.Problem = "go.mod or go.sum differ from what `go mod tidy` writes"
		f.Detail = firstLines(stdout, tidyDiffLines)
		f.Fix = &protocol.Fix{
			Description: "run go mod tidy in the Go module",
			AppliesTo:   []string{p.Name},
			Steps: []protocol.FixStep{{
				Description: "go mod tidy in " + p.Root + " (may use the network)",
				Argv:        []string{"go", "mod", "tidy"},
				Dir:         p.Root,
			}},
		}
	default:
		f.OK = true
		f.Detail = "could not check: " + untidyReason(stderr, err)
	}
	return f, true
}

// untidyReason says in one line why `go mod tidy -diff` gave no answer.
func untidyReason(stderr string, err error) string {
	switch {
	case errors.Is(err, toolrun.ErrTimeout):
		return "go mod tidy -diff took longer than " + tidyTimeout.String()
	case strings.Contains(stderr, "flag provided but not defined") || strings.Contains(stderr, "unknown flag"):
		return "go mod tidy -diff needs Go 1.23 or newer"
	case strings.Contains(stderr, "GOPROXY=off") || strings.Contains(stderr, "module lookup disabled") ||
		strings.Contains(stderr, "cannot find module") || strings.Contains(stderr, "missing go.sum"):
		return "offline: modules are not in the module cache"
	}
	if msg := firstLines(strings.TrimSpace(stderr), 1); msg != "" {
		return msg
	}
	return err.Error()
}

// firstLines returns the first n lines of s, with a note when it cuts.
func firstLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\r\n"), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:n], "\n") + "\n... and " + strconv.Itoa(len(lines)-n) + " more lines"
}

// CheckModulesTidy runs CheckModuleTidy for every project, a few at a time, and
// returns the findings in the order of the projects.
func CheckModulesTidy(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	found := make([]*protocol.Finding, len(projects))
	var wg sync.WaitGroup
	slots := make(chan struct{}, tidyParallel)
	for i, p := range projects {
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			if f, ok := CheckModuleTidy(workspaceRoot, p); ok {
				found[i] = &f
			}
		}()
	}
	wg.Wait()
	out := []protocol.Finding{}
	for _, f := range found {
		if f != nil {
			out = append(out, *f)
		}
	}
	return out
}
