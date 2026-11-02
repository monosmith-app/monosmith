package operation

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// CommitCommand is `smith commit`: it commits what is changed in the work tree
// as one operation, after the fact, for the person who did not tick the commit
// box before a change (smith never commits unasked).
//
// Returns: the command.
func CommitCommand() *cobra.Command {
	var yes, dryRun bool
	cmd := &cobra.Command{
		Use:   "commit [operation words]",
		Short: "Commit the changes in the work tree as one operation that smith undo can revert",
		Long: `Commit every change of the work tree (smith's own .smith folder left out) as one
commit named after an operation, such as "add app web", and record it so that smith undo can
revert it. It is what ticking "commit" before smith add does, for a change that is already in
the files.

It lists the files first and asks in a terminal. With no terminal, or in CI, it refuses
unless you pass --yes. Because it takes everything that git status shows, check the list: a
change of your own in the same work tree is committed with it.`,
		Example: "  smith commit add app web --dry-run\n  smith commit add app web",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if command == "" {
				command = "save the changes"
			}
			plan := operationplan.Plan{Title: "commit " + command}
			for _, l := range lines {
				plan.Add("Goes into one commit (chore: "+command+")", l)
			}
			proceed, err := operationplan.Present(out, cmd.InOrStdin(), plan, operationplan.Options{
				Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment(),
			})
			if err != nil || !proceed {
				return err
			}
			hash, err := Always(root).Finish(command)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Committed %s: chore: %s (smith undo reverts it)\n", short(hash), command)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "list the files that would be committed and change nothing")
	return cmd
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
