package agentdoc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// FileCommand is `smith agent file`.
func FileCommand(version string) *cobra.Command {
	var check, noClaude bool
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Write AGENTS.md (and a CLAUDE.md that points to it) from the workspace's rules",
		Long: `Write AGENTS.md from the rules of the workspace, for the AI coding agents that
work in it: where code goes (the architecture in force), how files are named,
how to add a project, which commands mean done, which derived files not to edit
by hand, and the choices of each project.

The generated text sits between the lines <!-- smith:begin agent-instructions -->
and <!-- smith:end agent-instructions -->. Write your own text anywhere else in
the file: a new run keeps it. A file with no markers keeps its text and gets the
generated region at the end. The text is the same for the same workspace.

CLAUDE.md gets a short region that imports AGENTS.md. --no-claude-md skips it.

--check writes nothing and fails when a file is missing or out of date, for CI.`,
		Example: "  smith agent file\n  smith agent file --check\n  smith agent file --no-claude-md",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			facts, err := Collect(s)
			if err != nil {
				return err
			}
			changes, err := PlanFiles(s.Root, facts, !noClaude)
			var broken *BrokenFileError
			if errors.As(err, &broken) {
				return problem.New(broken.Error(), "keep one <!-- smith:begin agent-instructions --> line and one <!-- smith:end agent-instructions --> line after it, or delete both")
			}
			if err != nil {
				return err
			}
			return finish(cmd, s.Root, changes, check)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "write nothing; fail when a file is missing or out of date")
	cmd.Flags().BoolVar(&noClaude, "no-claude-md", false, "do not write CLAUDE.md")
	return cmd
}

// finish reports the changes and, unless check is set, writes them.
func finish(cmd *cobra.Command, root string, changes []FileChange, check bool) error {
	out := cmd.OutOrStdout()
	var stale, missing []string
	for _, c := range changes {
		if !c.Differs {
			fmt.Fprintf(out, "%s is current\n", c.Path)
			continue
		}
		if c.Exists {
			stale = append(stale, c.Path)
		} else {
			missing = append(missing, c.Path)
		}
		if check {
			if c.Exists {
				fmt.Fprintf(out, "%s is out of date\n", c.Path)
			} else {
				fmt.Fprintf(out, "%s is missing, run smith agent file\n", c.Path)
			}
			continue
		}
		if err := os.WriteFile(filepath.Join(root, c.Path), []byte(c.Content), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "wrote %s\n", c.Path)
	}
	if !check {
		return nil
	}
	var parts []string
	if len(missing) > 0 {
		parts = append(parts, listWithVerb(missing)+" missing")
	}
	if len(stale) > 0 {
		parts = append(parts, listWithVerb(stale)+" not up to date with the workspace")
	}
	if len(parts) > 0 {
		return problem.New(strings.Join(parts, "; "), "smith agent file")
	}
	return nil
}

// listWithVerb joins paths with "and" and adds "is" or "are".
func listWithVerb(paths []string) string {
	if len(paths) > 1 {
		return strings.Join(paths, " and ") + " are"
	}
	return paths[0] + " is"
}
