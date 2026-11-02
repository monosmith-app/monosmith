package scaffold

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// firstCommitFlags is the flag line that names the first commit choice.
const firstCommitFlags = "--first-commit smith|all|none   (what the first commit holds: only smith's files, everything, or nothing; --no-commit is the same as none)"

// RefuseOpenNewChoices is the refusal of smith new when nobody can answer: no
// terminal and no --yes. Every open decision (the CI, the main language, the
// code layout, the agent files, the dev container and the first commit) is
// named with its flag in one message, so the person answers all at once. A
// workspace is never written with choices smith made silently.
//
// Parameters: answers are the questions the flags and the recipe answered;
// firstCommit is --first-commit and noCommit is --no-commit.
// Returns: the problem, or nil when every decision has an answer.
func RefuseOpenNewChoices(answers map[string]string, firstCommit string, noCommit bool) error {
	var flags []string
	for _, q := range NewQuestions() {
		if q.Advanced || q.AtEnd || q.Flag == "" || len(q.Options) == 0 {
			continue
		}
		if q.Applies != nil && !q.Applies(answers) {
			continue
		}
		if _, given := answers[q.Key]; given {
			continue
		}
		values := make([]string, len(q.Options))
		for i, o := range q.Options {
			values[i] = o.Value
		}
		flags = append(flags, fmt.Sprintf("--%s %s   (default %s)", q.Flag, strings.Join(values, "|"), q.Default))
	}
	if firstCommit == "" && !noCommit {
		flags = append(flags, firstCommitFlags)
	}
	if len(flags) == 0 {
		return nil
	}
	return problem.New("nobody can answer the questions of smith new: there is no terminal and --yes was not given",
		append(flags, "--yes takes the default of each and prints it", "--dry-run shows the plan and changes nothing")...)
}

// FirstCommitText says what the first commit of smith new holds, for the plan
// and the confirmation.
//
// Parameters: name is the workspace name; mode is the CommitDecision mode;
// noCommit is true when nothing is committed.
// Returns: one sentence that starts with "First commit:".
func FirstCommitText(name, mode string, noCommit bool) string {
	subject := "chore: create the " + name + " workspace with smith"
	switch {
	case noCommit || mode == CommitNone:
		return "First commit: none. smith writes its files and commits nothing (--first-commit smith commits them)."
	case mode == CommitAll:
		return "First commit: " + subject + ", with every file in the folder (--first-commit smith commits only smith's files)."
	default:
		return "First commit: " + subject + ", with only the files smith writes (--first-commit none or --no-commit commits nothing)."
	}
}

// PlanNew lists the files smith new would create, without creating them in
// the real folder: it builds the workspace in a temporary folder and reads
// the file names, so the list is the real one.
//
// Parameters: opts are the settled options of the command.
// Returns: the plan, titled with the workspace name; an error when smith could
// not build the temporary copy.
func PlanNew(opts NewOptions) (operationplan.Plan, error) {
	tmp, err := os.MkdirTemp("", "smith-new-plan-")
	if err != nil {
		return operationplan.Plan{}, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	copyOpts := opts
	copyOpts.Dir, copyOpts.NoCommit, copyOpts.Out = filepath.Join(tmp, opts.Name), true, io.Discard
	copyOpts.SaveRecipe = ""
	if _, err := NewWorkspace(copyOpts); err != nil {
		return operationplan.Plan{}, err
	}
	var files []string
	err = filepath.WalkDir(copyOpts.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(copyOpts.Dir, path)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return operationplan.Plan{}, err
	}
	sort.Strings(files)
	plan := operationplan.Plan{Title: "create the " + opts.Name + " workspace"}
	for _, f := range files {
		plan.Add("Files smith creates", operationplan.Line{Action: operationplan.Create, Path: f})
	}
	return plan, nil
}
