package release

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/releasepreview"
	"github.com/monosmith-app/monosmith/apps/smith/releaserehearsal"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
	"github.com/spf13/cobra"
)

// Command is `smith release`.
func Command(version string) *cobra.Command {
	var (
		dryRun, yes, noPush, asJSON, comment, notes, status, verbose, rehearse bool
		format                                                                 string
		hold, bumpFlags                                                        []string
	)
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Version, publish and tag every package with releasable commits",
		Long: `Release every publishable package from its conventional commits since its last tag:
feat → minor, fix/perf → patch, a breaking change → major (minor before 1.0).
Packages release in dependency order. Each is stamped, built, published and
tagged; manifests are restored afterwards (the tag is the record). A re-run
after a failure resumes where it stopped. Before the first publish every
npm, PyPI, NuGet and RubyGems registry is asked whether it accepts the
credential and the names, so a bad key stops the release before anything is
published. What was published where is kept in dist/release/<project>/journal.json
(smith release --status). smith release --rehearse runs registry checks for every
package on their own, before a first release, and publishes nothing.

A branch other than the release branch can release a line of prerelease
versions of its own (2.0.0-beta.1, beta.2, ...): declare it as a
[[release.channels]] table in smith.toml (smith config set does not edit lists
of tables). docs/reference/releasing.md, Release channels, has an example.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if comment && !dryRun {
				return errors.New("--comment posts a preview: use it with --dry-run")
			}
			preview := func(markdown string) {
				if comment {
					releasepreview.Post(markdown, cmd.ErrOrStderr())
				}
			}
			if verbose {
				// The plugins read it: a preview shows one line per tool unless asked.
				_ = os.Setenv(toolrun.VerboseEnv, "1")
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err == nil {
				// Tags and branch are read once for the whole command.
				s.Git = s.Git.WithRefCache()
			}
			if err != nil {
				return err
			}
			defer s.Close()
			if status {
				journals, err := ReadJournals(s.Root)
				if err != nil {
					return err
				}
				if asJSON {
					enc := json.NewEncoder(out)
					enc.SetIndent("", "  ")
					return enc.Encode(journals)
				}
				fmt.Fprint(out, RenderJournalStatus(journals))
				return nil
			}
			if rehearse {
				return rehearseRegistries(s, out, asJSON, verbose)
			}
			var blockers []string
			for _, f := range CheckLineage(s, true) {
				if f.Blocking {
					blockers = append(blockers, f.Sentence())
				} else {
					fmt.Fprintln(cmd.ErrOrStderr(), "warning: "+f.Sentence())
				}
			}
			if len(blockers) > 0 {
				preview("### Release preview\n\n**Blocked:** the release tags would give wrong versions, so there is no plan:\n\n- " + strings.Join(blockers, "\n- ") + "\n")
				return fmt.Errorf("cannot release: the release tags would give wrong versions, so no plan is shown:\n  - %s", strings.Join(blockers, "\n  - "))
			}
			plan, err := PlanFor(s)
			if err != nil {
				return err
			}
			bumps, err := ParseBumps(bumpFlags)
			if err != nil {
				return err
			}
			plan, adjusted, err := Adjust(plan, Adjustments{Hold: hold, Bump: bumps})
			if err != nil {
				return err
			}
			if !asJSON {
				for _, n := range adjusted {
					fmt.Fprintln(cmd.ErrOrStderr(), "note: "+n)
				}
			}
			if dryRun {
				// A real release refuses these in its preflight: the preview says so too.
				if problems := ChannelRefusals(s, plan); len(problems) > 0 {
					preview("### Release preview\n\n**Blocked:** this release would be refused:\n\n- " + strings.Join(problems, "\n- ") + "\n")
					return fmt.Errorf("cannot release:\n  - %s", strings.Join(problems, "\n  - "))
				}
			}
			if err := CheckAPI(cmd.Context(), s, &plan); err != nil {
				preview("### Release preview\n\n**Blocked:** " + err.Error() + "\n")
				return err
			}
			if len(plan.Items) == 0 {
				// A re-run on a fresh checkout has no journal: a release that
				// stopped after its tag is finished from the tag.
				plan.Items = UnfinishedFromGitHub(s, cmd.ErrOrStderr())
			}
			if !asJSON {
				if printsTextPlan(asJSON, format) {
					printPlan(cmd, plan, dryRun, verbose)
				}
				if notes {
					for _, item := range plan.Items {
						fmt.Fprintf(out, "\n# %s %s\n\n%s", item.Project, item.Next, RenderNotes(item))
					}
				}
			}
			if len(plan.Items) == 0 {
				if dryRun {
					preview(Report{DryRun: true, Channel: plan.Channel, Unchanged: plan.Unchanged}.Markdown())
				}
				if asJSON {
					return json.NewEncoder(out).Encode(Report{DryRun: dryRun, Channel: plan.Channel, Unchanged: plan.Unchanged, Results: []ReportResult{}})
				}
				return nil
			}
			if !dryRun && !yes {
				if !operationplan.ShouldAsk(operationplan.DetectEnvironment(), false) {
					return errors.New("refusing to release without a terminal; pass --yes (CI does). Hold a package back with --hold <project>, change a bump with --bump <project>=patch|minor|major, or see the plan with --dry-run")
				}
				in := bufio.NewReader(os.Stdin)
				var go_ bool
				plan, go_, err = Review(in, out, plan, func(p Plan) { printPlan(cmd, p, false, verbose) }, Adjustments{})
				if err != nil {
					return err
				}
				if !go_ || len(plan.Items) == 0 {
					return errors.New("release cancelled: nothing was tagged or published")
				}
				fmt.Fprintln(out, Consequences(plan, !noPush))
				fmt.Fprint(out, "Release these packages? [y/N] ")
				answer, _ := in.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(answer), "y") {
					return errors.New("release cancelled: nothing was tagged or published")
				}
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			report, err := Execute(ctx, s, plan, Options{DryRun: dryRun, Push: !noPush, Out: out, Terse: format != "markdown"})
			if dryRun {
				preview(report.Markdown())
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			} else {
				fmt.Fprintln(out)
				if format == "markdown" {
					fmt.Fprint(out, report.Markdown())
				} else {
					fmt.Fprint(out, report.Text())
				}
				if err != nil {
					// The report above says "Stopped: <why>": do not print it again.
					return ErrReported
				}
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be released; change nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().StringSliceVar(&hold, "hold", nil, "do not release this package this time (repeat it, or separate names with commas); a package that releases only because of it is held back too")
	cmd.Flags().StringSliceVar(&bumpFlags, "bump", nil, "release a package with this bump instead of the one its commits ask for: <project>=patch|minor|major")
	cmd.Flags().BoolVar(&noPush, "no-push", false, "create tags locally only")
	cmd.Flags().BoolVar(&notes, "notes", false, "also print each release's notes (what its GitHub Release says)")
	cmd.Flags().BoolVar(&status, "status", false, "show the release journal: what was published where (dist/release/<project>/journal.json)")
	cmd.Flags().BoolVar(&rehearse, "rehearse", false, "check every registry the release publishes to (it resolves, which credential it expects, that yours is accepted, that the CI identity may publish) with read-only requests; publish nothing")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the release report as JSON")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "name every unchanged package")
	cmd.Flags().StringVar(&format, "format", "text", "how the result is printed: text, or markdown (what the CI job summary and --comment use)")
	cmd.Flags().BoolVar(&comment, "comment", false, "with --dry-run in a pull request build, post the preview on the pull request (GitHub, Azure Repos), updating smith's earlier comment")
	return cmd
}

// rehearseRegistries is `smith release --rehearse`: it asks every registry
// the workspace's packages publish to and fails when one would refuse.
func rehearseRegistries(s *session.Session, out io.Writer, asJSON, verbose bool) error {
	targets := releaserehearsal.ListTargets(s, os.Getenv)
	report := releaserehearsal.Rehearse(targets, releaserehearsal.NewProbe(), os.Getenv)
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		fmt.Fprint(out, releaserehearsal.RenderText(report, len(targets), verbose))
	}
	if report.Failed > 0 {
		return fmt.Errorf("the rehearsal found %d registry problem(s): a real release would fail", report.Failed)
	}
	return nil
}

// printsTextPlan is whether the text plan comes before the result: not for JSON,
// and not for markdown, which is the whole result by itself (a job summary or
// a pull request comment).
func printsTextPlan(asJSON bool, format string) bool {
	return !asJSON && format != "markdown"
}

func printPlan(cmd *cobra.Command, plan Plan, dryRun, verbose bool) {
	out := cmd.OutOrStdout()
	if dryRun {
		fmt.Fprintln(out, "Dry run: nothing is published or tagged.")
	}
	if plan.Channel != "" {
		fmt.Fprintln(out, ChannelNotice(plan.Channel))
	}
	if len(plan.Items) == 0 {
		fmt.Fprintln(out, NothingToRelease())
	}
	for _, item := range plan.Items {
		from := item.Current
		if from == "" {
			from = "unreleased"
		}
		fmt.Fprintf(out, "  %-20s %s → %s  (%s)\n", item.Project, from, item.Next, item.Reason)
		for _, w := range item.APIWarnings {
			fmt.Fprintf(out, "  warning: %s\n", w)
		}
	}
	if len(plan.Unchanged) > 0 {
		if verbose {
			fmt.Fprintf(out, "Unchanged: %s\n", strings.Join(plan.Unchanged, ", "))
		} else {
			fmt.Fprintf(out, "Unchanged: %d packages (--verbose names them)\n", len(plan.Unchanged))
		}
	}
}
