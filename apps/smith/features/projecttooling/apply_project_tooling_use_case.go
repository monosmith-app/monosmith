package projecttooling

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/features/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/features/fullconversion"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

// Input is the one decision Apply takes: which tools the project at Dir, in
// the workspace at Root, uses.
type Input struct {
	// Root is the workspace folder (absolute).
	Root string
	// Dir is the project folder, workspace-relative ("." at the root).
	Dir string
	// Answer is Keep or Smith.
	Answer Answer
}

// RunOptions say how Apply asks and checks.
type RunOptions struct {
	Out io.Writer
	// In reads the answers to the removal questions; used only when Interactive.
	In          *bufio.Reader
	Interactive bool
	// Yes removes every group of old tools without asking.
	Yes bool
	// Smith is the command that runs smith for the install and the checks.
	Smith string
	// SkipFinish leaves out the install and the checks (a caller that applies
	// several projects runs them once, with the last).
	SkipFinish bool
	// Verify builds, tests, lints and runs doctor after the install.
	Verify bool
	// Files lists the workspace's files, workspace-relative; nil reads git.
	Files func() ([]string, error)
	// WholeRoot puts every file at the workspace root in the removal plan's
	// scope (see InScope): smith adopt sets it, smith import does not.
	WholeRoot bool
}

// Result is what Apply did.
type Result struct {
	// Name is the project's name.
	Name string
	// Lint and Build are true when smith's tool is in place.
	Lint, Build bool
	// Applied and Declined are the removal groups, by ID.
	Applied, Declined []string
	// Left are the steps smith could not take, each with its fix.
	Left []string
	// Findings are the checks of the finish that failed.
	Findings []doctorfix.Finding
}

// Apply takes the answer for one Node project. Keep changes nothing. Smith
// first puts smith's lint (PlaceSmithLint) and build (PlaceSmithBuild) in
// place, then retires the old tools with the full conversion's removal plan,
// limited to the project's folder and the workspace root, and only for the
// tool whose replacement is in place: a lint group when the lint is ready, a
// build group when the build is ready. Each group is asked (or applied with
// Yes). It ends with the install and the checks, unless SkipFinish. The
// lint check runs without ESLint's fixes: smith reports what its rules find in
// the code it took in, and does not rewrite that code unasked.
//
// `smith adopt` calls it for each project of a repository, and `smith import`
// for the project it brings in.
//
// Parameters: ctx stops the checks; in is the decision; o the run options.
// Returns: what was done and what is left; an error for a file that cannot be
// read or written, or a removal that failed.
func Apply(ctx context.Context, in Input, o RunOptions) (Result, error) {
	in.Dir = cleanRel(in.Dir)
	a, err := Assess(in.Root, in.Dir)
	if err != nil {
		return Result{}, err
	}
	res := Result{Name: a.Name}
	if in.Answer != Smith || !a.Node {
		return res, nil
	}
	fmt.Fprintf(o.Out, "\nTooling of %s: smith's build and ESLint config go in, then the old tools go\n", a.Name)
	lint, err := PlaceSmithLint(in.Root, in.Dir, o.Out)
	if err != nil {
		return res, err
	}
	build, err := PlaceSmithBuild(in.Root, a, o.Out)
	if err != nil {
		return res, err
	}
	res.Lint, res.Build = lint.Ready, build.Ready
	res.Left = append(append(res.Left, lint.Left...), build.Left...)
	choices := convertchoices.Choices{Level: convertchoices.LevelFull, Lint: convertchoices.Keep, Test: convertchoices.Keep, Build: convertchoices.Keep, CI: convertchoices.Smith}
	if lint.Ready {
		choices.Lint = convertchoices.Smith
	}
	if build.Ready {
		choices.Build = convertchoices.Smith
	}
	files := o.Files
	if files == nil {
		files = gitclient.New(in.Root).Files
	}
	run, err := fullconversion.Run(ctx, fullconversion.RunOptions{
		Root: in.Root, Out: o.Out, In: o.In, Interactive: o.Interactive, Yes: o.Yes,
		Choices: choices, SkipFinish: o.SkipFinish, Verify: o.Verify, NoLintFix: true, Smith: o.Smith,
		Files: func() ([]string, error) {
			all, err := files()
			if err != nil {
				return nil, err
			}
			return InScope(all, in.Dir, o.WholeRoot), nil
		},
	})
	if err != nil {
		return res, err
	}
	res.Applied, res.Declined, res.Findings = run.Applied, run.Declined, run.Findings
	for _, id := range run.Declined {
		res.Left = append(res.Left, fmt.Sprintf("%s keeps the old tools of the group %q, as you chose; remove them by hand when you are ready", a.Name, id))
	}
	return res, nil
}

// InScope keeps the files the step of one project may change: the files in
// the project's folder, the root package.json and the root ESLint flat
// config. With wholeRoot every file at the workspace root is in scope too:
// smith adopt moved the repository's own tool files (an .eslintrc) there,
// and they are the project's. The other projects of the workspace are out of
// scope, so their tools stay.
//
// Parameters: files are workspace-relative paths; dir is the project folder;
// wholeRoot puts every root file in scope.
// Returns: the files in scope, in their order.
func InScope(files []string, dir string, wholeRoot bool) []string {
	if dir == "." {
		return files
	}
	var out []string
	prefix := strings.TrimSuffix(dir, "/") + "/"
	for _, f := range files {
		atRoot := path.Dir(f) == "."
		switch {
		case strings.HasPrefix(f, prefix):
		case atRoot && (wholeRoot || f == "package.json" || slices.Contains(rootFlatConfigs, f)):
		default:
			continue
		}
		out = append(out, f)
	}
	return out
}
