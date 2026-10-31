package why

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith why`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "why <file>",
		Short: "Say who owns a file, what produced it and how to change it (why-miss says why a task reran)",
		Long: `Say who owns a file, what produced it and how to change it.

For a file that smith derives, why names the owner (smith or you), the renderer
and its version, the facts and project graph it follows from, and its state.
Then it says how to change it: edit the input and run smith derive, or run
smith derive --eject <file> to own the file.

For a file that a smith command generates (CI pipelines, AGENTS.md, the agent
files), why names the command. For a file in a project that no record claims, it
says you own it, and names the stack and the choices of the project, because a
plugin may have scaffolded the file once. For a file that smith does not know,
it says so.

The path is relative to the current directory. --json prints the same answer for
tools. smith why-miss is a different command: it says why a task ran again.`,
		Example: "  smith why packages/web/tsconfig.lib.json\n  smith why .github/workflows/ci.yml\n  smith why AGENTS.md --json",
		Args:    commandargs.Exactly(commandargs.Spec{Names: []string{"a file"}, Example: "smith why packages/web/tsconfig.json", More: "smith why-miss <project>:<task> says why a task reran"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			path, err := relative(s.Root, args[0])
			if err != nil {
				return err
			}
			rec, err := derivelock.Read(s.Root)
			if err != nil {
				return err
			}
			in := Inputs{Record: rec, Graph: s.Graph, Stacks: s.Config.Stacks, Head: headReader(s.Root)}
			if s.Config.Derive.Enabled {
				plan, _, err := derive.BuildPlan(s)
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "warning: the state of derived files is not shown:", err)
				} else {
					in.Plan = &plan
				}
			}
			e := Explain(in, path)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(e)
			}
			Print(cmd.OutOrStdout(), e)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}

// relative turns a path given relative to the current directory into a
// workspace-relative path with forward slashes. It returns a problem when the
// path lies outside the workspace.
func relative(root, arg string) (string, error) {
	abs, err := filepath.Abs(arg)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", problem.New(arg+" is outside the workspace", "give a path inside "+root)
	}
	return filepath.ToSlash(rel), nil
}

// headReader reads the first 2 KiB of a workspace file.
func headReader(root string) func(string) string {
	return func(path string) string {
		f, err := os.Open(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return ""
		}
		defer func() { _ = f.Close() }()
		buf := make([]byte, 2048)
		n, _ := f.Read(buf)
		return string(buf[:n])
	}
}
