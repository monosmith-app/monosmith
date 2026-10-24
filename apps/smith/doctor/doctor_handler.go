package doctor

import (
	"os"
	"strconv"
	"strings"

	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// ErrUnhealthy makes doctor exit non-zero.
var ErrUnhealthy = errors.New("the workspace has problems")

// Command is `smith doctor`.
func Command(version string) *cobra.Command {
	var asJSON, fix, yes, verbose bool
	var only []string
	var choose []int
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the workspace's health; every problem names its remedy",
		Long: `Check the workspace's health: smith's invariants and each plugin's checks.
Every problem names its remedy. With --fix, doctor shows the fixes it can make:
file changes as diffs, commands it would run (npm install, git rm --cached),
and questions it would ask (install the missing tool, or remove the script?).
The same fix on many projects is one fix that says which it applies to.

A safe fix only restores what smith or a plugin owns, like go.work or a
generated pipeline with its slots kept; doctor applies those once you confirm
(or with --yes). Any other fix, and any with a question, is asked about in a
terminal. Without a terminal it is shown with the command that applies it:
--only "<check or fix>" --yes applies a fix you name (and doctor exits non-zero
while a named fix is not applied), with --choose 1,2 answering its questions by
option number. Steps run in the workspace root and stop at the first that fails.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if (len(only) > 0 || len(choose) > 0) && !fix {
				return errors.New("--only and --choose pick what to fix, so they go with --fix")
			}
			if fix && asJSON {
				return errors.New("--json prints the report and its fixes; apply them with --fix, without --json")
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			report := Run(s)
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			} else {
				var names []string
				for _, p := range s.Graph.Projects {
					names = append(names, p.Name, p.Root)
				}
				for _, f := range report.Findings {
					if f.OK && !verbose {
						continue
					}
					mark := "✔"
					if !f.OK {
						mark = "✖"
						if f.Warning {
							mark = "!"
						}
					}
					fmt.Fprintf(out, "%s %s\n", mark, f.Check)
					if !f.OK {
						fmt.Fprintf(out, "    %s\n", f.Detail)
						for _, step := range strings.Split(f.Remedy, "\n") {
							fmt.Fprintf(out, "    → %s\n", step)
						}
					}
				}
				if !verbose {
					for _, line := range CollapsePassed(report.Findings, names) {
						fmt.Fprintf(out, "✔ %s\n", line.Text)
					}
				}
				fmt.Fprintf(out, "\n%d passed, %d failed, %d warning(s)\n", report.Passed, report.Failed, report.Warnings)
				if !fix && len(report.Fixes) > 0 {
					fmt.Fprintln(out, fixesAvailable(report.Fixes))
				}
			}
			if fix {
				if report, err = fixAll(cmd, s, report, yes, only, choose, func() (*session.Session, error) {
					return session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
				}); err != nil {
					return err
				}
			}
			if report.Failed > 0 {
				return ErrUnhealthy
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	cmd.Flags().BoolVar(&fix, "fix", false, "show the fixes doctor can make as diffs, and apply the safe ones")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "list every passed check, not one line for each check that repeats per project")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "with --fix, apply without asking")
	cmd.Flags().IntSliceVar(&choose, "choose", nil, "with --fix and --only, answer the fix's questions by option number, in order (1,2)")
	cmd.Flags().StringArrayVar(&only, "only", nil, "with --fix, fix only the named check, as doctor prints it (repeatable)")
	return cmd
}

// fixesAvailable tells how many fixes `doctor --fix` offers and how many ask.
func fixesAvailable(plan []doctorfix.Planned) string {
	asking := 0
	for _, p := range plan {
		if p.Ask {
			asking++
		}
	}
	text := fmt.Sprintf("%d fix(es) available: `smith doctor --fix`", len(plan))
	if asking > 0 {
		text += fmt.Sprintf(" (%d will ask you first)", asking)
	}
	return text
}

// quoted puts each name in double quotes.
func quoted(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = fmt.Sprintf("%q", n)
	}
	return out
}

// ErrFixesPending makes doctor --fix --only exit non-zero when the named fix
// needs a decision nobody could give (no terminal): the output names the
// command that applies it.
var ErrFixesPending = errors.New("some fixes need your decision: run the command shown under each")

// fixRounds is how many times doctor plans again after a fix made another
// one's file stale (two fixes that edit smith.toml): the second is made anew
// from what the first left.
const fixRounds = 3

// fixAll walks the fixes the findings carry (see doctorfix.Walk), then checks
// again from a fresh session, since fixes change smith.toml and package.json.
func fixAll(cmd *cobra.Command, s *session.Session, report Report, yes bool, only []string, choose []int, reopen func() (*session.Session, error)) (Report, error) {
	out := cmd.OutOrStdout()
	plan := Plan(Only(report, only))
	if len(plan) == 0 && len(only) > 0 {
		fmt.Fprintf(out, "\nNo failed check or fix named %s carries a fix.\n", strings.Join(quoted(only), ", "))
		return report, nil
	}
	if len(plan) == 0 {
		fmt.Fprintln(out, "\nNothing doctor can fix itself: follow the remedies above.")
		return report, nil
	}
	if len(choose) > 0 {
		asking := 0
		for _, p := range plan {
			if len(p.Choices) > 0 {
				asking++
			}
		}
		switch {
		case asking == 0:
			return report, errors.New("--choose answers a fix's questions, and the fix named has none")
		case asking > 1:
			return report, errors.New("--choose answers the questions of one fix: name it with --only")
		}
	}
	interactive := term.IsTerminal(int(os.Stdin.Fd()))
	var applied, pending int
	var opened *session.Session // the session a replan opened; the caller closes the first
	defer func() {
		if opened != nil {
			opened.Close()
		}
	}()
	done := map[string]bool{} // fixes declined or left pending: not offered again
	for round := 0; round < fixRounds && len(plan) > 0; round++ {
		fmt.Fprintln(out)
		res, err := doctorfix.Walk(cmd.Context(), plan, doctorfix.Options{
			Root: s.Root, Out: out, In: cmd.InOrStdin(), Interactive: interactive,
			Yes: yes, Explicit: len(only) > 0, Choose: choose,
			Command: func(p doctorfix.Planned, answers []int) string { return retryCommand(p, answers) },
		})
		for _, p := range res.Pending {
			done[p.Fix.Description] = true
			pending++
			fmt.Fprintf(out, "\nnot applied: %s needs your decision. Run one of:\n", p.Fix.Description)
			for _, c := range p.Commands {
				fmt.Fprintf(out, "  %s\n", c)
			}
		}
		for _, d := range res.Skipped {
			done[d] = true
		}
		applied += len(res.Applied)
		if err != nil {
			return report, err
		}
		if len(res.Stale) == 0 || len(res.Applied) == 0 {
			break
		}
		// A fix was made from a file another fix then changed: look again.
		fresh, err := reopen()
		if err != nil {
			return report, err
		}
		if opened != nil {
			opened.Close()
		}
		opened, s = fresh, fresh
		report = Run(s)
		plan = nil
		for _, p := range Plan(Only(report, only)) {
			if !done[p.Description] {
				plan = append(plan, p)
			}
		}
	}
	// Fixes left pending fail the command when the user named them; otherwise
	// the exit code is the workspace's health, as without --fix.
	unresolved := pending > 0 && len(only) > 0
	if applied == 0 {
		if unresolved {
			return report, ErrFixesPending
		}
		return report, nil
	}
	fresh, err := reopen()
	if err != nil {
		return report, err
	}
	defer fresh.Close()
	after := Run(fresh)
	fmt.Fprintf(out, "\nafter fixing: %d passed, %d failed, %d warning(s)\n", after.Passed, after.Failed, after.Warnings)
	if unresolved {
		return after, ErrFixesPending
	}
	return after, nil
}

// retryCommand is the command line that applies a fix with the given answers.
func retryCommand(p doctorfix.Planned, answers []int) string {
	line := fmt.Sprintf("smith doctor --fix --only %q --yes", p.Description)
	if len(answers) > 0 {
		parts := make([]string, len(answers))
		for i, a := range answers {
			parts[i] = strconv.Itoa(a)
		}
		line += " --choose " + strings.Join(parts, ",")
	}
	return line
}
