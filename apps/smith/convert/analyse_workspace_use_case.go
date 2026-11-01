package convert

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/fullconversion"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Analysis is an Nx or Turborepo workspace read for conversion: what the
// command does before it asks, available to a caller that asks in its own
// way (the TUI's adopt wizard).
type Analysis struct {
	Root    string
	Version string
	Plan    *convertplan.Plan
	Git     *gitclient.Git
	// Files are the tracked files, as git lists them.
	Files []string
	// Source names the kind of workspace: "Nx workspace", "Turborepo workspace".
	Source string
	// Blocked is true when the plan has findings that stop the conversion;
	// Verdict says which.
	Blocked bool
	Verdict string
	// Counts reads "53 projects (1 app, 52 packages) and 13 it cannot convert":
	// the words the command line report uses, for a screen to show as they are,
	// so both say the same. Blockers is how many findings stop the conversion.
	Counts   string
	Blockers int
	// Dirty is true when the working tree has uncommitted changes.
	Dirty bool
}

// Analyse reads the workspace at root and plans its conversion; nothing is
// changed. It fails with a problem that says what to do when the folder is
// already a smith workspace, is not a git repository, or is not an Nx or
// Turborepo workspace.
func Analyse(root, version string) (*Analysis, error) {
	return AnalyseWithProgress(root, version, nil)
}

// AnalyseWithProgress is Analyse that tells progress each stage of the read
// ("Asking Nx to describe the projects…", "Planned 53 projects."), so a screen
// can show what is happening during the minutes a large repository takes. A
// nil progress is ignored.
func AnalyseWithProgress(root, version string, progress Progress) (*Analysis, error) {
	if _, err := os.Stat(filepath.Join(root, workspace.FileName)); err == nil {
		return nil, problem.New("this is already a smith workspace ("+workspace.FileName+" exists)", "smith upgrade brings it up to the installed smith")
	}
	git := gitclient.New(root)
	if !git.IsRepository() {
		return nil, problem.New("this folder is not a git repository: convert makes its changes through git, so each one can be reviewed and undone", `git init && git add -A && git commit -m "before smith", then run this again`)
	}
	files, err := git.Files()
	if err != nil {
		return nil, err
	}
	tags, _ := git.Tags("")
	input := convertplan.Input{
		DirName:      filepath.Base(root),
		Files:        files,
		Tags:         tags,
		SmithVersion: version,
		Read: func(rel string) string {
			raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			return string(raw)
		},
	}
	plan, err := readPlan(root, files, input, progress)
	if err != nil {
		return nil, err
	}
	a := &Analysis{
		Root: root, Version: version, Plan: plan, Git: git, Files: files,
		Source: sourceLabel(plan), Blocked: plan.Blocked(false), Verdict: Verdict(plan, false),
		Counts: projectCount(plan), Blockers: plan.Count(convertplan.Blocker),
	}
	if status, err := git.Run("status", "--porcelain"); err == nil && DirtyApartFromInstall(status) {
		a.Dirty = true
	}
	return a, nil
}

// MakeView applies the conversion on a scratch copy and returns it, so the
// plan of the removals can be made without touching the workspace. The
// caller closes it.
func (a *Analysis) MakeView() (*fullconversion.View, error) {
	return fullconversion.MakeView(a.Root, a.Files, func(dir string) error {
		scratch := *a.Plan
		scratch.Root = dir
		_, err := Apply(&scratch, ApplyOptions{Version: a.Version, SkipInstall: true, Out: io.Discard})
		return err
	})
}

// Convert applies the conversion to the workspace without installing: the
// caller runs the install and the checks. keepPipelines leaves the CI
// pipelines as they are. It returns the summary the command prints.
func (a *Analysis) Convert(keepPipelines bool, log io.Writer) (string, error) {
	if a.Blocked {
		return "", ErrBlocked
	}
	if a.Dirty {
		return "", errors.New("the working tree has uncommitted changes: the conversion would not be one reviewable change")
	}
	res, err := Apply(a.Plan, ApplyOptions{Version: a.Version, SkipInstall: true, KeepPipelines: keepPipelines, Out: log})
	if err != nil {
		return "", err
	}
	if err := saveReport(a.Root, Report(a.Plan, ReportOptions{Verbose: true})); err != nil {
		return "", err
	}
	staged := stageNewFiles(a.Git, a.Plan, res)
	return Summary(a.Plan, res.LostSlots, staged), nil
}

// SmithCommand is how a check runs smith: by name when the smith on the path
// is this one, else by the path of the running executable.
func SmithCommand() string { return smithCommand() }
