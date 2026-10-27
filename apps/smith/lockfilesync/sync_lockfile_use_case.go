package lockfilesync

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Runner runs npm with args in dir and returns what it printed, both streams
// together.
type Runner func(dir string, args ...string) (string, error)

// Options say what to sync and where.
type Options struct {
	Root string
	// Out is where a failure and a rebuilt lockfile are told. Nothing else is
	// written: npm's own output stays hidden unless a step fails.
	Out io.Writer
	// Create installs even when there is no package-lock.json yet, as long as
	// no other package manager's lockfile is there. Without it a workspace
	// that has no package-lock.json is left alone.
	Create bool
	// LockOnly updates package-lock.json without touching node_modules, for
	// the callers that did not ask for an install.
	LockOnly bool
	// Run runs npm; the installed npm when nil.
	Run Runner
	// Tracked tells whether git tracks package-lock.json, so that rebuilding
	// it can be taken back; git ls-files when nil.
	Tracked func(root string) bool
}

// Result tells what Sync did.
type Result struct {
	// Skipped is why nothing ran ("" when something did).
	Skipped string
	// Regenerated is true when npm could not fix the lockfile and smith
	// rebuilt it from scratch.
	Regenerated bool
	// Notes are npm's warnings, folded to one line each.
	Notes []string
}

var otherLockfiles = []string{"yarn.lock", "pnpm-lock.yaml", "bun.lock", "bun.lockb"}

// Sync runs npm install so package-lock.json matches package.json, and checks
// that npm ci accepts the result. npm 11 can reuse the installed tree on the
// first run and leave the lockfile out of step, so a second install
// re-resolves it when the check fails. When that still fails and git tracks
// the lockfile, it is deleted and made again (git can take that back);
// otherwise nothing is deleted and the error says what to do. node_modules is
// never deleted.
func Sync(o Options) (Result, error) {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(o.Root, rel))
		return err == nil
	}
	if !exists("package.json") {
		return Result{Skipped: "there is no package.json"}, nil
	}
	if !exists("package-lock.json") {
		if !o.Create {
			return Result{Skipped: "there is no package-lock.json"}, nil
		}
		for _, other := range otherLockfiles {
			if exists(other) {
				return Result{Skipped: "the workspace uses " + other + ", not npm"}, nil
			}
		}
	}
	run := o.Run
	if run == nil {
		if _, err := exec.LookPath("npm"); err != nil {
			fmt.Fprintln(o.Out, "! npm is not installed: run npm install to update package-lock.json")
			return Result{Skipped: "npm is not installed"}, nil
		}
		run = runNpm
	}
	tracked := o.Tracked
	if tracked == nil {
		tracked = gitTracksLockfile
	}
	install := []string{"install", "--no-audit", "--no-fund"}
	if o.LockOnly {
		install = append(install, "--package-lock-only")
	}
	check := []string{"ci", "--dry-run", "--ignore-scripts", "--no-audit", "--no-fund"}

	// attempt installs then checks; it returns the install's warnings, or the
	// output of the step that failed and whether a second try can help (the
	// check failed, not the install).
	attempt := func() (warnings, failed string, retry bool, err error) {
		installed, err := run(o.Root, install...)
		if err != nil {
			return "", installed, false, err
		}
		if checked, err := run(o.Root, check...); err != nil {
			return "", checked, true, err
		}
		return installed, "", false, nil
	}

	var failed string
	for try := 1; try <= 2; try++ {
		warnings, out, retry, err := attempt()
		if err == nil {
			return Result{Notes: toolrun.FoldWarnings(warnings, "run npm install to see them")}, nil
		}
		failed = out
		if !retry {
			break
		}
	}

	lock := filepath.Join(o.Root, "package-lock.json")
	original, readErr := os.ReadFile(lock)
	if readErr != nil || !tracked(o.Root) {
		fmt.Fprint(o.Out, failureTail(failed))
		return Result{}, errors.New("package-lock.json is out of step with package.json and npm could not fix it; run npm install until `npm ci --dry-run` passes, then commit package-lock.json")
	}
	if err := os.Remove(lock); err != nil {
		return Result{}, err
	}
	warnings, failed, _, err := attempt()
	if err != nil {
		_ = os.WriteFile(lock, original, 0o644)
		fmt.Fprint(o.Out, failureTail(failed))
		return Result{}, errors.New("package-lock.json is out of step with package.json and npm could not rebuild it either; smith put your package-lock.json back. Fix what npm reports above, then run npm install")
	}
	fmt.Fprintln(o.Out, "! npm could not bring package-lock.json in step with package.json, so smith deleted it and made it again. Package versions may have moved: `git diff package-lock.json` shows how, and `git checkout package-lock.json` takes it back.")
	return Result{Regenerated: true, Notes: toolrun.FoldWarnings(warnings, "run npm install to see them")}, nil
}

func runNpm(dir string, args ...string) (string, error) {
	var log bytes.Buffer
	cmd := exec.Command("npm", args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = dir, &log, &log
	err := cmd.Run()
	return log.String(), err
}

func gitTracksLockfile(root string) bool {
	_, err := gitclient.New(root).Run("ls-files", "--error-unmatch", "package-lock.json")
	return err == nil
}
