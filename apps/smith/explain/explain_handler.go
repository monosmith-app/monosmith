package explain

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith explain`.
func Command(version string) *cobra.Command {
	var asJSON, recipeOnly bool
	var onlyTask string
	cmd := &cobra.Command{
		Use:   "explain <project>[:<task>]",
		Short: "Show how a project's tasks resolved, and where each setting came from (why-miss says why one reran)",
		Long: `Show how a project's tasks resolved, and where each of their fields came from:
the language plugin, a stack's plugin, a convention, or smith.toml.

For a project made from a recipe, smith explain <project> also lists what the
recipe decided for it (its choices, the plugin at each level and the plugins'
settings), each with the recipe file and line it came from. --recipe prints
only that, and with --json prints it for tools. The tasks' --json output is
unchanged.

The text starts with one line per task (its command, whether it is cached,
what it waits for), then each task's fields. Where a value came from is said
once for all of them, and marked only where it differs. --task build, or
<project>:build, shows one task.`,
		Example: "  smith explain web\n  smith explain web:build\n  smith explain web --recipe",
		Args:    commandargs.Exactly(commandargs.Spec{Names: []string{"a project name"}, Example: "smith explain web", More: "smith projects lists them; add :<task> for one task, e.g. smith explain web:build"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			out := cmd.OutOrStdout()
			target := args[0]
			if onlyTask != "" && !strings.Contains(target, ":") {
				target += ":" + onlyTask
			}
			project, task, _ := strings.Cut(target, ":")
			if recipeOnly || (task == "" && !asJSON) {
				e, ok, err := recipedrift.Explain(s, project)
				switch {
				case recipeOnly && err != nil:
					return err
				case recipeOnly && !ok:
					return fmt.Errorf("%s was not made from a recipe: smith.toml records none for it", project)
				case err != nil:
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: the recipe of %v\n", err)
				case ok && asJSON:
					return encode(out, e)
				case ok:
					printRecipe(out, e)
				}
				if recipeOnly {
					return nil
				}
			}
			tasks, err := Explain(s.Graph, target)
			if err != nil {
				return err
			}
			if asJSON {
				return encode(out, tasks)
			}
			printTasks(out, tasks)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	cmd.Flags().StringVar(&onlyTask, "task", "", "explain only this task (same as <project>:<task>)")
	cmd.Flags().BoolVar(&recipeOnly, "recipe", false, "print only what the project's recipe decided, and where")
	return cmd
}

func encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
