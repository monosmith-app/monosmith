package kindchange

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/spf13/cobra"
)

// report is what --json prints.
type report struct {
	Plan
	DryRun  bool    `json:"dryRun"`
	Outcome Outcome `json:"outcome"`
}

// Command is `smith kind`.
func Command(version string) *cobra.Command {
	var opts Options
	var asJSON, allowDirty bool
	cmd := &cobra.Command{
		Use:   "kind <project> <kind>",
		Short: "Make a project an app, an internal lib or a published package",
		Long: `Make a project another kind after it was created: an app (deployed), a lib
(internal, shipped inside what uses it) or a package (published to a registry).

smith asks the project's language plugin what the change involves, and does all
of it: the kind in the manifest, the guard against publishing (dropped for a
package, put back for a lib or an app), the folder moved from libs/ to
packages/ to apps/ with git so its history follows, every place that names the
folder (workspace members, path dependencies, imports of a module whose path
ends in the folder, go.work, smith.toml), and the CI pipelines regenerated.

It refuses, naming the reason, when the change would break something: an app
that other projects import, a published package that others import, a package
that was already released. --force goes ahead against those. Files smith does
not know how to rewrite (a Dockerfile, a script, a document) are listed after
the change, for you to check.

Use --dry-run first: it lists every file that would be edited and the move,
and changes nothing. Without it smith shows the same plan, with a diff of each
file, and asks before it writes; without a terminal it needs --yes. The work
tree must be clean (--allow-dirty overrides). Nothing is committed unless you
pass --commit.`,
		Example: "  smith kind core package --dry-run\n  smith kind core package\n  smith kind core lib",
		Args: commandargs.Exactly(commandargs.Spec{
			Names:   []string{"the project to change", "the kind it becomes: app, lib or package"},
			Example: "smith kind core package --dry-run",
			More:    "smith projects lists the projects and their kinds",
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			switch kind := protocol.ProjectKind(strings.ToLower(args[1])); kind {
			case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
				opts.To = kind
			default:
				return problem.New(fmt.Sprintf("%q is not a kind of project", args[1]), "the kinds are app, lib and package")
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			plan, err := PlanChange(s, opts)
			if err != nil {
				return err
			}
			rep := report{Plan: plan, DryRun: opts.DryRun}
			if !opts.DryRun {
				if !allowDirty {
					if dirty := DirtyPaths(statusOf(s)); len(dirty) > 0 {
						return problem.New("the work tree has uncommitted changes, and the move would mix with them ("+strings.Join(firstN(dirty, 3), ", ")+")",
							"commit or stash them, then run this again", "--allow-dirty goes ahead anyway")
					}
				}
				yes, _ := cmd.Flags().GetBool("yes")
				shown := cmd.OutOrStdout()
				if asJSON {
					shown = cmd.ErrOrStderr()
				}
				if _, err := operationplan.Present(shown, cmd.InOrStdin(), OperationPlan(s.Root, plan), operationplan.Options{Yes: yes, Ask: true, Env: operationplan.DetectEnvironment()}); err != nil {
					return err
				}
				if rep.Outcome, err = Apply(s, plan); err != nil {
					return err
				}
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(rep)
			}
			print(out, rep)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "list every file edit and the move; change nothing")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "go ahead although other projects import it or it was already released")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the plan and what was done as JSON")
	cmd.Flags().BoolVar(&allowDirty, "allow-dirty", false, "go ahead although the work tree has uncommitted changes")
	return operation.Wrap(cmd, operation.Spec{AddYes: true, Describe: func(_ *cobra.Command, args []string) string {
		return "make " + args[0] + " " + args[1]
	}})
}

// statusOf is git status --porcelain of the workspace; "" when git cannot say.
func statusOf(s *session.Session) string {
	out, _ := s.Git.Run("status", "--porcelain")
	return out
}

// firstN is at most n of the items.
func firstN(items []string, n int) []string {
	if len(items) > n {
		return items[:n]
	}
	return items
}

func print(out io.Writer, r report) {
	verb := "Made"
	if r.DryRun {
		verb = "Would make"
	}
	fmt.Fprintf(out, "%s %s %s (it was %s)\n", verb, r.Project, article(r.To), article(r.From))
	for _, why := range r.Overruled {
		fmt.Fprintf(out, "  going ahead although %s\n", why)
	}
	if r.Moves() {
		how := "git mv"
		if !r.DryRun && r.Outcome.Moved != "" {
			how = r.Outcome.Moved
		}
		fmt.Fprintf(out, "  moves   %s -> %s (%s)\n", r.Root, r.NewRoot, how)
	} else {
		fmt.Fprintf(out, "  stays in %s: it is not in a libs/, packages/ or apps/ folder\n", r.Root)
	}
	for _, e := range r.Edits {
		fmt.Fprintf(out, "  edits   %s\n", e.Path)
	}
	if r.DryRun {
		fmt.Fprintln(out, "  the CI pipelines would be regenerated")
	}
	for _, f := range r.Outcome.Regenerated {
		fmt.Fprintf(out, "  writes  %s\n", f)
	}
	if len(r.Warnings) > 0 || len(r.Outcome.Notes) > 0 {
		fmt.Fprintln(out, "Check:")
		for _, w := range append(append([]string{}, r.Warnings...), r.Outcome.Notes...) {
			fmt.Fprintf(out, "  - %s\n", w)
		}
	}
	if len(r.Mentions) > 0 {
		fmt.Fprintf(out, "Still name %s, which smith does not rewrite:\n", r.Root)
		const shown = 12
		for i, f := range r.Mentions {
			if i == shown {
				fmt.Fprintf(out, "  - and %d more\n", len(r.Mentions)-shown)
				break
			}
			fmt.Fprintf(out, "  - %s\n", f)
		}
	}
}
