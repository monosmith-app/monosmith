package release

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/releasepreview"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Command is `smith release`.
func Command(version string) *cobra.Command {
	var (
		dryRun, yes, noPush, asJSON, comment, notes, status bool
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
(smith release --status).`,
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
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
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
			plan, err := MakePlan(s.Git, s.Graph)
			if err != nil {
				return err
			}
			if !asJSON {
				printPlan(cmd, plan)
				if notes {
					for _, item := range plan.Items {
						fmt.Fprintf(out, "\n# %s %s\n\n%s", item.Project, item.Next, RenderNotes(item))
					}
				}
			}
			if len(plan.Items) == 0 {
				if dryRun {
					preview(Report{DryRun: true, Unchanged: plan.Unchanged}.Markdown())
				}
				if asJSON {
					return json.NewEncoder(out).Encode(Report{DryRun: dryRun, Unchanged: plan.Unchanged, Results: []ReportResult{}})
				}
				return nil
			}
			if !dryRun && !yes {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					return errors.New("refusing to release without a terminal; pass --yes (CI does)")
				}
				fmt.Fprint(out, "Release these packages? [y/N] ")
				answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(answer), "y") {
					return errors.New("release cancelled")
				}
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			report, err := Execute(ctx, s, plan, Options{DryRun: dryRun, Push: !noPush, Out: out})
			if dryRun {
				preview(report.Markdown())
			}
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			} else {
				fmt.Fprintln(out)
				fmt.Fprint(out, report.Markdown())
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be released; change nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().BoolVar(&noPush, "no-push", false, "create tags locally only")
	cmd.Flags().BoolVar(&notes, "notes", false, "also print each release's notes (what its GitHub Release says)")
	cmd.Flags().BoolVar(&status, "status", false, "show the release journal: what was published where (dist/release/<project>/journal.json)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the release report as JSON")
	cmd.Flags().BoolVar(&comment, "comment", false, "with --dry-run in a pull request build, post the preview on the pull request (GitHub, Azure Repos), updating smith's earlier comment")
	return cmd
}

func printPlan(cmd *cobra.Command, plan Plan) {
	out := cmd.OutOrStdout()
	if len(plan.Items) == 0 {
		fmt.Fprintln(out, NothingToRelease())
	}
	for _, item := range plan.Items {
		from := item.Current
		if from == "" {
			from = "unreleased"
		}
		fmt.Fprintf(out, "  %-20s %s → %s  (%s)\n", item.Project, from, item.Next, item.Reason)
	}
	if len(plan.Unchanged) > 0 {
		fmt.Fprintf(out, "  unchanged: %s\n", strings.Join(plan.Unchanged, ", "))
	}
}
