package eslintfix

import (
	"fmt"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

// Target is one project ESLint fixes and lints.
type Target struct {
	// Name is the project's name, as `smith run lint -p` takes it.
	Name string
	// Dir is the project's absolute folder; ESLint runs there.
	Dir string
}

// Outcome is what ESLint left in one project.
type Outcome struct {
	Project string
	// Fixed is how many problems ESLint fixed.
	Fixed int
	// Errors and Warnings are the problems ESLint could not fix.
	Errors, Warnings int
	// Crash is the first line of ESLint's error when it could not run at all
	// ("" when it ran); Remedy says what to do about it.
	Crash, Remedy string
}

// Problems is the count of problems left: errors and warnings.
func (o Outcome) Problems() int { return o.Errors + o.Warnings }

// Runner runs argv in dir and returns its standard output, its error output
// and its exit code. err is set only when the program could not start.
type Runner func(dir string, argv []string) (stdout, stderr string, code int, err error)

// Options say where to fix and how to run ESLint.
type Options struct {
	// Root is the workspace root, whose node_modules holds ESLint.
	Root    string
	Targets []Target
	Run     Runner
	// ESLintVersion is the version of the installed ESLint, "" when none is
	// installed.
	ESLintVersion string
}

// Fix lints each target in its own folder with the workspace's own ESLint
// (each project's eslint.config.* is found from its folder): once to count,
// then with --fix. A crash in one project does not stop the others. It
// returns what ESLint fixed and left per project, in the targets' order.
//
// Parameters: o says where and how. Returns one Outcome per target.
func Fix(o Options) []Outcome {
	out := make([]Outcome, 0, len(o.Targets))
	bin := filepath.Join(o.Root, "node_modules", "eslint", "bin", "eslint.js")
	for _, t := range o.Targets {
		res := Outcome{Project: t.Name}
		if o.ESLintVersion == "" {
			res.Crash = "ESLint is not installed in the workspace"
			res.Remedy = "run npm install, then smith run lint"
			out = append(out, res)
			continue
		}
		// Two runs: the problems before the fixes, then the fixes and what
		// they leave; the difference is what ESLint fixed.
		before, crash := lint(o.Run, t.Dir, []string{"node", bin, ".", "--format", "json"})
		if crash == "" {
			var after Outcome
			after, crash = lint(o.Run, t.Dir, []string{"node", bin, ".", "--fix", "--format", "json"})
			res.Errors, res.Warnings = after.Errors, after.Warnings
			res.Fixed = max(0, before.Problems()-after.Problems())
		}
		if crash != "" {
			res.Crash, res.Remedy = crash, remedy(t.Name, o.ESLintVersion)
		}
		out = append(out, res)
	}
	return out
}

// lint runs ESLint once and returns the problems it reports, or crash: the
// first line of its error when it could not run.
func lint(run Runner, dir string, argv []string) (counted Outcome, crash string) {
	stdout, stderr, code, err := run(dir, argv)
	if err != nil {
		return counted, err.Error()
	}
	errors, warnings, parsed := countProblems(stdout)
	if code > 1 || !parsed {
		if crash = firstErrorLine(stderr + "\n" + stdout); crash == "" {
			crash = fmt.Sprintf("ESLint exited with code %d", code)
		}
		return counted, crash
	}
	counted.Errors, counted.Warnings = errors, warnings
	return counted, ""
}

// remedy is what to do when ESLint crashed in project: raise an ESLint older
// than smith's config needs, else read the whole error.
func remedy(project, installed string) string {
	if !removalplan.ESLintMeetsConfig(installed) {
		return fmt.Sprintf("smith's config needs eslint %s and the workspace has %s: set \"eslint\": %q in package.json, then run npm install", removalplan.LintESLintRange, installed, removalplan.LintESLintRange)
	}
	return "run " + LintCommand(project) + " to see the whole error"
}
