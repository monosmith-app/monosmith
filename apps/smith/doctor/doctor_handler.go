package doctor

import (
	"bufio"
	"os"
	"strings"

	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// ErrUnhealthy makes doctor exit non-zero.
var ErrUnhealthy = errors.New("the workspace has problems")

// Command is `smith doctor`.
func Command(version string) *cobra.Command {
	var asJSON, fix, yes, verbose bool
	var only []string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the workspace's health; every problem names its remedy",
		Long: `Check the workspace's health: smith's invariants and each plugin's checks.
Every problem names its remedy. With --fix, doctor shows the fixes it can make
as diffs, then applies the safe ones once you confirm (or with --yes): a safe
fix only restores what smith or a plugin owns, like go.work or a generated
pipeline with its slots kept. Others are shown, never applied. With --only, doctor
takes up just the fix of the named check (the check's text, as doctor prints it).`,
		Example: `  smith doctor
  smith doctor --fix        # review the diffs, then confirm
  smith doctor --fix --yes
  smith doctor --fix --only "go.work exists" --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(only) > 0 && !fix {
				return errors.New("--only names the checks to fix, so it goes with --fix")
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
			}
			if fix {
				if report, err = fixAll(cmd, s, report, yes, only); err != nil {
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
	cmd.Flags().StringArrayVar(&only, "only", nil, "with --fix, fix only the named check, as doctor prints it (repeatable)")
	return cmd
}

// quoted puts each name in double quotes.
func quoted(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = fmt.Sprintf("%q", n)
	}
	return out
}

// fixAll shows every fix the findings carry, applies the safe ones once
// confirmed, and checks again.
func fixAll(cmd *cobra.Command, s *session.Session, report Report, yes bool, only []string) (Report, error) {
	out := cmd.OutOrStdout()
	safe, manual := Fixes(Only(report, only))
	if len(safe)+len(manual) == 0 && len(only) > 0 {
		fmt.Fprintf(out, "\nNo failed check named %s carries a fix.\n", strings.Join(quoted(only), ", "))
		return report, nil
	}
	if len(safe)+len(manual) == 0 {
		fmt.Fprintln(out, "\nNothing doctor can fix itself: follow the remedies above.")
		return report, nil
	}
	fmt.Fprintln(out)
	for _, f := range manual {
		fmt.Fprintln(out, "not applied (it could lose something of yours; apply it by hand):")
		if err := ShowFix(out, s.Root, f); err != nil {
			return report, err
		}
	}
	if len(safe) == 0 {
		return report, nil
	}
	for _, f := range safe {
		if err := ShowFix(out, s.Root, f); err != nil {
			return report, err
		}
	}
	if !yes {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Fprintln(out, "\nPass --yes to apply these fixes.")
			return report, nil
		}
		fmt.Fprintf(out, "\nApply %d fix(es)? [y/N] ", len(safe))
		answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			fmt.Fprintln(out, "Nothing applied.")
			return report, nil
		}
	}
	for _, f := range safe {
		if err := ApplyFix(s.Root, f); err != nil {
			return report, err
		}
		fmt.Fprintf(out, "✔ %s\n", f.Description)
	}
	after := Run(s)
	fmt.Fprintf(out, "\nafter fixing: %d passed, %d failed, %d warning(s)\n", after.Passed, after.Failed, after.Warnings)
	return after, nil
}
