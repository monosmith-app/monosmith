package operation

import (
	"bufio"
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// The scopes of a commit after the fact.
const (
	scopeAll = "all"
	scopeOp  = "operation"
)

// CommitCommand is `smith commit`: it commits what is changed in the work tree
// as one operation, after the fact, for the person who did not tick the commit
// box before a change (smith never commits unasked). When smith knows which
// files the last operation changed and the work tree holds other changes, the
// person chooses: only the operation's files, or everything git shows.
//
// Returns: the command.
func CommitCommand() *cobra.Command {
	var yes, dryRun, onlyOperation, everything bool
	cmd := &cobra.Command{
		Use:   "commit [operation words]",
		Short: "Commit the changes of an operation as one commit that smith undo can revert",
		Long: `Commit the changes of the last operation smith left uncommitted (smith add, remove,
upgrade...) as one commit named after it, such as "add app web", and record it so that smith
undo can revert it. It is what ticking "commit" before smith add does, for a change that is
already in the files.

smith remembers which files the operation changed. When the work tree holds other changes of
your own, it asks: commit only the files of the operation (the default), everything git shows
changed, or cancel. Without a terminal it refuses and names the answers: --only-operation or
--all. With nothing else changed, it commits the operation's files without asking this.
--dry-run lists the files and commits nothing; --yes skips the final confirmation only.`,
		Example: "  smith commit --dry-run\n  smith commit --only-operation --yes\n  smith commit add app web --all",
		RunE: func(cmd *cobra.Command, args []string) error {
			if onlyOperation && everything {
				return problem.New("--only-operation and --all say opposite things", "give one of them")
			}
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			git := gitclient.New(root)
			if !git.IsRepository() {
				return problem.New("this workspace is not in a git work tree", "run git init, then commit again")
			}
			if !git.HasCommits() {
				return problem.New("the repository has no commit yet", "make the first commit yourself, then smith can commit the changes after it")
			}
			lines, err := changedLines(git)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(lines) == 0 {
				fmt.Fprintln(out, "Nothing to commit: the work tree has no change.")
				return nil
			}
			command := strings.TrimSpace(strings.Join(args, " "))
			journal, known := ReadPending(root)
			if command == "" && known {
				command = journal.Command
			}
			if command == "" {
				command = "save the changes"
			}
			var own, extra []operationplan.Line
			for _, l := range lines {
				if known && slices.Contains(journal.Files, l.Path) {
					own = append(own, l)
				} else {
					extra = append(extra, l)
				}
			}
			scope, err := chooseScope(cmd, known && len(own) > 0, own, extra, onlyOperation, everything, yes, dryRun)
			if err != nil || scope == "" {
				return err
			}
			chosen := lines
			if scope == scopeOp {
				chosen = own
			}
			plan := operationplan.Plan{Title: "commit " + command}
			for _, l := range chosen {
				plan.Add("Goes into one commit (chore: "+command+")", l)
			}
			proceed, err := operationplan.Present(out, cmd.InOrStdin(), plan, operationplan.Options{
				Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment(),
			})
			if err != nil || !proceed {
				return err
			}
			var hash string
			if scope == scopeOp {
				paths := make([]string, len(chosen))
				for i, l := range chosen {
					paths[i] = l.Path
				}
				hash, err = Always(root).FinishOnly(command, paths)
			} else {
				hash, err = Always(root).Finish(command)
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Committed %s: chore: %s (smith undo reverts it)\n", short(hash), command)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for the final confirmation (the choice of files still needs --only-operation or --all)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "list the files that would be committed and change nothing")
	cmd.Flags().BoolVar(&onlyOperation, "only-operation", false, "commit only the files the last operation changed; the other changes stay in the work tree")
	cmd.Flags().BoolVar(&everything, "all", false, "commit everything git shows as changed, your own changes included")
	return cmd
}

// chooseScope decides which files go into the commit. With nothing else changed,
// or no record of the operation, there is nothing to choose. Otherwise a flag
// answers, or a terminal asks; with neither it refuses and names the flags. It
// returns "" with no error when a dry run listed the two sets, or the person cancelled.
func chooseScope(cmd *cobra.Command, known bool, own, extra []operationplan.Line, onlyOperation, everything, yes, dryRun bool) (string, error) {
	switch {
	case everything:
		return scopeAll, nil
	case onlyOperation && !known:
		return "", problem.New("smith has no record of the files of the last operation here", "--all commits everything git shows as changed")
	case onlyOperation:
		return scopeOp, nil
	case !known || len(extra) == 0:
		return scopeAll, nil
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "The last operation changed %d file(s). Git shows %d more that it did not change:\n", len(own), len(extra))
	for _, l := range extra {
		fmt.Fprintf(out, "  %s\n", l.Path)
	}
	env := operationplan.DetectEnvironment()
	switch {
	case dryRun:
		fmt.Fprintln(out, "--only-operation commits the first set, --all both. Dry run: nothing was changed.")
		return "", nil
	case yes || !env.Terminal || env.CI:
		return "", problem.New("smith cannot choose which files to commit: your own changes are in the work tree too",
			"--only-operation commits only the files of the operation", "--all commits everything git shows as changed")
	}
	fmt.Fprintf(out, "  1) only the %d file(s) this operation changed (default)\n  2) everything git shows changed (%d files)\n  3) cancel\nChoose [1]: ", len(own), len(own)+len(extra))
	// One reader for both questions, so the second does not lose what the first buffered.
	in := bufio.NewReader(cmd.InOrStdin())
	cmd.SetIn(in)
	answer, _ := in.ReadString('\n')
	switch strings.TrimSpace(answer) {
	case "", "1":
		return scopeOp, nil
	case "2":
		return scopeAll, nil
	}
	fmt.Fprintln(out, "Cancelled: nothing was committed.")
	return "", nil
}

// changedLines are the changes of the work tree as plan lines, each file on
// its own line, smith's .smith folder left out.
func changedLines(git *gitclient.Git) ([]operationplan.Line, error) {
	status, err := git.Run("status", "--porcelain", "-uall", "--", ".", exclude)
	if err != nil {
		return nil, err
	}
	var lines []operationplan.Line
	for _, row := range strings.Split(status, "\n") {
		if len(row) < 4 {
			continue
		}
		code, path := row[:2], row[3:]
		if _, to, renamed := strings.Cut(path, " -> "); renamed {
			path = to
		}
		path = strings.Trim(path, `"`)
		action := operationplan.Change
		switch {
		case code == "??" || strings.Contains(code, "A"):
			action = operationplan.Create
		case strings.Contains(code, "D"):
			action = operationplan.Remove
		}
		lines = append(lines, operationplan.Line{Action: action, Path: path})
	}
	return lines, nil
}
