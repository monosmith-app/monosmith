package derive

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith derive`.
func Command(version string) *cobra.Command {
	var check bool
	var explain, adopt, eject, replace, reclaim []string
	cmd := &cobra.Command{
		Use:   "derive",
		Short: "Write the files that follow from facts and the project graph, and report hand edits",
		Long: `Write the files that follow from facts and the project graph.

Some config repeats in every project and follows from what smith knows. Two
things are derived:

  - what the project graph says: a tsconfig's references are the project's
    internal dependencies, and a Jest displayName (between smith:begin and
    smith:end markers) is the project's name;
  - the shape of the sibling projects: when at least three projects of one kind,
    and two thirds of those that have the file, hold the same config file
    (tsconfig*.json, .swcrc, jest.config.*, and the like) after the project name
    and the references are substituted, that shape is derived for each project
    that has the file. A key or a file that differs is the project's own: smith
    leaves it alone and --explain says it diverges from its siblings.

A plugin renders such a file, or part of it, and smith keeps a record of what it
wrote in smith.lock: the owner (smith or you), the renderer and a hash of the
content. Delete a derived file and smith derive writes it back.

A file smith owns is rewritten when its inputs change. When its content no
longer matches the record, someone edited it: derive reports the drift with a
diff and never overwrites it. Then you choose, per file:

  --replace <file>  let smith overwrite it with the derived content
  --adopt <file>    keep the edit as the new baseline (a file smith owns by
                    keys, that still has every derived key)
  --eject <file>    own the file: the record says "you" and smith stops
  --reclaim <file>  the opposite of --eject, in one step: give an ejected or
                    drifted file back to smith. It shows the diff of what will
                    be overwritten, asks (or --yes), writes the derived content
                    and records smith as the owner again. --dry-run shows the
                    diff only; --commit commits the change for smith undo.

A file that exists, is not in the record and differs from the derived content is
left alone in the same way (--replace takes it over, --eject keeps it).

--check writes nothing and fails when derive would change a file or the record,
or when a file needs a choice. smith ci verify runs it. --explain <file> says
who owns a file, which renderer wrote it and why it is in its state.

Derived files are opt-in: set [derive] enabled = true in smith.toml.`,
		Example: "  smith derive\n  smith derive --check\n  smith derive --explain packages/web/tsconfig.lib.json\n  smith derive --replace packages/web/tsconfig.lib.json\n  smith derive --eject packages/web/tsconfig.lib.json\n  smith derive --reclaim packages/web/tsconfig.lib.json --commit",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if !s.Config.Derive.Enabled && check {
				fmt.Fprintln(cmd.OutOrStdout(), "Derived files are off in this workspace ([derive] enabled is not true): nothing to check.")
				return nil
			}
			if !s.Config.Derive.Enabled {
				return problem.New("derived files are off in this workspace", "set enabled = true under [derive] in smith.toml")
			}
			yes, _ := cmd.Flags().GetBool("yes")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			return run(cmd.OutOrStdout(), cmd.InOrStdin(), s, check, yes, dryRun, explain, Decision{Replace: replace, Adopt: adopt, Eject: eject, Reclaim: reclaim})
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "write nothing; fail when a derived file or the record is out of date")
	cmd.Flags().StringArrayVar(&explain, "explain", nil, "say who owns this file and why it is in its state")
	cmd.Flags().StringArrayVar(&replace, "replace", nil, "overwrite this drifted or diverged file with the derived content")
	cmd.Flags().StringArrayVar(&adopt, "adopt", nil, "take the hand edit of this keys-mode file as the new baseline")
	cmd.Flags().StringArrayVar(&reclaim, "reclaim", nil, "give this ejected or drifted file back to smith: shows the diff, asks, rewrites it and records smith as the owner")
	cmd.Flags().StringArrayVar(&eject, "eject", nil, "own this file: smith stops touching it")
	return operation.Wrap(cmd, operation.Spec{AddYes: true, AddDryRun: true, Describe: func(*cobra.Command, []string) string { return "derive files" }})
}

func run(out io.Writer, in io.Reader, s *session.Session, check, yes, dryRun bool, explain []string, d Decision) error {
	plan, rec, err := BuildPlan(s)
	if err != nil {
		return err
	}
	if len(explain) > 0 {
		for _, p := range explain {
			items := itemsAt(plan, p)
			if d := plan.FindDivergence(p); len(items) == 0 && d != nil {
				fmt.Fprintf(out, "%s\n  owner     you\n  renderer  %s (plugin %s)\n  project   %s\n  state     %s\n", d.Path, d.Renderer, d.Plugin, d.Project, d.Detail)
				continue
			}
			if len(items) == 0 {
				return problem.New("smith does not derive "+p, "smith derive --check lists the files it derives")
			}
			PrintExplanation(out, items)
		}
		return nil
	}
	if check {
		return CheckPlan(out, plan)
	}
	writes := PlanWrites(s.Root, plan)
	if len(d.Reclaim) > 0 {
		changes, err := PlanReclaim(s.Root, plan, d.Reclaim)
		if err != nil {
			return err
		}
		PrintReclaimDiffs(out, d.Reclaim, changes)
		fmt.Fprintln(out)
		writes = ReclaimPlan(d.Reclaim, changes)
	}
	proceed, err := operationplan.Present(out, in, writes, operationplan.Options{
		Yes: yes, DryRun: dryRun, Ask: len(d.Reclaim) > 0 || len(d.Replace)+len(d.Adopt)+len(d.Eject) == 0, Env: operationplan.DetectEnvironment(),
	})
	if err != nil || !proceed {
		return err
	}
	before := rec
	before.Files = append([]derivelock.Entry(nil), rec.Files...)
	outcome, err := Apply(s.Root, plan, &rec, d)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(before.Files, rec.Files) {
		if err := derivelock.Write(s.Root, rec); err != nil {
			return err
		}
	}
	printOutcome(out, plan, outcome)
	if len(outcome.Held) > 0 {
		PrintDiffs(out, outcome.Held)
		return problem.New(fmt.Sprintf("%d derived files need a choice: smith did not overwrite them", len(outcome.Held)),
			"--replace <file> overwrites with the derived content", "--adopt <file> keeps a hand edit of a keys-mode file", "--eject <file> makes the file yours")
	}
	return nil
}

// CheckPlan reports the plan to out and fails when anything needs action: the
// check behind smith derive --check and smith ci verify.
func CheckPlan(out io.Writer, plan Plan) error {
	pending := plan.Pending()
	if len(pending) == 0 {
		fmt.Fprintf(out, "%d derived files are up to date.\n", len(plan.Items))
		return nil
	}
	PrintPlan(out, plan)
	title := fmt.Sprintf("%d derived files are out of date", len(pending))
	for _, i := range pending {
		switch i.Status {
		case StatusDrift, StatusDiverged, StatusBroken:
			// smith derive would hold these back, so name the choices instead.
			return problem.New(title, "smith derive --explain <file> shows the diff",
				"then --replace, --adopt or --eject <file> picks who owns the file")
		}
	}
	return problem.New(title, "run smith derive, and commit the files and smith.lock")
}

func printOutcome(out io.Writer, plan Plan, o Outcome) {
	list := func(verb string, paths []string) {
		for _, p := range paths {
			fmt.Fprintf(out, "  %-9s %s\n", verb, p)
		}
	}
	list("wrote", o.Written)
	list("replaced", o.Replaced)
	list("recorded", o.Recorded)
	list("adopted", o.Adopted)
	list("ejected", o.Ejected)
	list("reclaimed", o.Reclaimed)
	for _, h := range o.Held {
		fmt.Fprintf(out, "  %-9s %s  (%s)  %s\n", string(h.Status), h.File.Path, h.File.Renderer, h.Detail)
	}
	changed := len(o.Written) + len(o.Replaced) + len(o.Recorded) + len(o.Adopted) + len(o.Ejected) + len(o.Reclaimed)
	fmt.Fprintf(out, "%d derived files: %s.\n", len(plan.Items), strings.Join(summary(changed, len(o.Held)), ", "))
}

func summary(changed, held int) []string {
	parts := []string{fmt.Sprintf("%d changed", changed)}
	if held > 0 {
		parts = append(parts, fmt.Sprintf("%d need a choice", held))
	}
	return parts
}
