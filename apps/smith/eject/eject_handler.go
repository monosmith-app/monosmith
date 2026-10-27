package eject

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
	"github.com/spf13/cobra"
)

// Command is `smith eject`.
func Command(version string) *cobra.Command {
	var all, force bool
	cmd := &cobra.Command{
		Use:   "eject [tool...]",
		Short: "Copy smith's default configuration of a tool into the repository, to own and edit it",
		Long: `Copy smith's default configuration of a tool into the repository.

smith runs lint, type checks and tests with defaults it ships, so a repository
needs no configuration of its own. When you want to change one, eject it: the
default is copied to the file the tool reads (ruff.toml, .golangci.yml, ...) in
the workspace root, and from then on that file replaces the default whole and
smith passes no configuration of its own.

Without arguments it lists the defaults that apply to this workspace. Name the
tools to eject, or pass --all. A file that exists already is left alone unless
--force replaces it.`,
		Example: "  smith eject\n  smith eject ruff\n  smith eject mypy pytest\n  smith eject --all",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			if err := ResetDefaults(root); err != nil {
				return err
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			out := cmd.OutOrStdout()
			available := Available(s.Root)
			if len(args) == 0 && !all {
				return list(out, available)
			}
			tools, err := choose(args, all, available)
			if err != nil {
				return err
			}
			for _, t := range tools {
				targets, err := Eject(s.Root, t, force, projectsUsing(s.Graph, t))
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "Copied the default of %s to %s: it is yours now, and smith passes no configuration to %s.\n", t.Name, strings.Join(targets, ", "), t.Name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "eject every default that applies")
	cmd.Flags().BoolVar(&force, "force", false, "replace a file that exists already")
	return cmd
}

// choose resolves the names on the command line, or all of them, to tools whose
// default applies.
func choose(names []string, all bool, available []Candidate) ([]tooldefaults.Tool, error) {
	var tools []tooldefaults.Tool
	if all {
		for _, c := range available {
			tools = append(tools, c.Tool)
		}
		if len(tools) == 0 {
			return nil, problem.New("no default applies to this workspace", "the repository configures its tools itself, or no project uses one")
		}
		return tools, nil
	}
	for _, n := range names {
		t, ok := tooldefaults.Lookup(n)
		if !ok {
			return nil, problem.New("smith has no default for "+n, "the tools with a default: "+strings.Join(toolNames(), ", "))
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func toolNames() []string {
	var out []string
	for _, t := range tooldefaults.Tools {
		out = append(out, t.Name)
	}
	return out
}

func list(out io.Writer, available []Candidate) error {
	if len(available) == 0 {
		_, err := fmt.Fprintln(out, "No default applies to this workspace: the repository configures its tools, or no project uses one.")
		return err
	}
	fmt.Fprintln(out, "Defaults smith passes to tools in this workspace (smith eject <tool> copies one to the repository):")
	for _, c := range available {
		note := ""
		if c.Exists {
			note = "  (" + c.Tool.Target + " exists already)"
		}
		target := c.Tool.Target
		if c.Tool.PerProject {
			target += " in each project"
		}
		fmt.Fprintf(out, "  %-14s %-7s -> %s%s\n", c.Tool.Name, c.Tool.Language, target, note)
	}
	return nil
}
