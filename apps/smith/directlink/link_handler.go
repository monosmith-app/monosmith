package directlink

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// Command is smith link: make the smith command start the program directly.
//
// Parameters: version is the running smith.
// Returns: the cobra command.
func Command(version string) *cobra.Command {
	var status, undo bool
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Make the smith command start the program directly: faster to start, less memory",
		Long: `A package manager (npm, pnpm, pip, dotnet) starts smith through a launcher of its own language,
which stays alive as long as smith runs: about 110 ms slower to start and about 40 MB more memory for
every smith process. smith link points the smith command at the program itself, whatever the manager.
It finds the smith entries on PATH (smith.cmd, a symbolic link, a launcher program), keeps the original
of each, and changes them.

On Windows the entries call a hard link of the program in ~/.smith/run, so the installed file is never the
one that runs and an update can replace it while smith runs in other terminals.

The first run of every version does this by itself and says so; the package manager writes its entries
again when it updates smith, so the next version links again. SMITH_NO_LINK=1 stops that, and smith link
--undo gives the entries back. smith link is safe to run again.`,
		Example: "  smith link\n  smith link --status\n  smith link --undo",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			o := ForThisProcess(version)
			out := cmd.OutOrStdout()
			switch {
			case status && undo:
				return problem.New("--status and --undo cannot be used together", "run one of them")
			case status:
				printStatus(out, Check(o))
				return nil
			case undo:
				return runUndo(out, o)
			}
			return runLink(out, o)
		},
	}
	cmd.Flags().BoolVar(&status, "status", false, "show how smith starts now and change nothing")
	cmd.Flags().BoolVar(&undo, "undo", false, "give the entries back as the package manager wrote them")
	return cmd
}

func runLink(out io.Writer, o Options) error {
	report, err := Link(o)
	switch {
	case errors.Is(err, ErrReplaced):
		return problem.New(err.Error())
	case err != nil:
		return problem.New("could not link smith: "+err.Error(), "check that "+o.Home+" can be written")
	}
	if len(report.Changed) == 0 && len(report.Already) == 0 && len(report.Skipped) == 0 {
		fmt.Fprintln(out, "No smith entry of this installation was found, so there is nothing to change (smith was started by its path or npx).")
		return nil
	}
	for _, e := range report.Changed {
		fmt.Fprintf(out, "linked   %s (%s)\n", e.Path, e.Kind)
	}
	for _, e := range report.Already {
		fmt.Fprintf(out, "direct   %s\n", e.Path)
	}
	for _, s := range report.Skipped {
		fmt.Fprintf(out, "skipped  %s\n", s)
	}
	if len(report.Changed) > 0 {
		fmt.Fprintf(out, "The smith command now starts %s directly. Open a new terminal command to use it; smith link --undo gives the entries back.\n", report.Target)
	}
	return nil
}

func runUndo(out io.Writer, o Options) error {
	restored, skipped, err := Undo(o)
	if err != nil {
		return problem.New("could not forget the link: "+err.Error(), "delete "+StatePath(o.Home))
	}
	for _, e := range restored {
		fmt.Fprintf(out, "restored %s\n", e.Path)
	}
	for _, s := range skipped {
		fmt.Fprintf(out, "skipped  %s\n", s)
	}
	if len(restored) == 0 && len(skipped) == 0 {
		fmt.Fprintln(out, "Nothing was linked.")
	}
	return nil
}

func printStatus(out io.Writer, s Status) {
	how := "through the launcher of its package manager"
	if s.Direct {
		how = "directly, with no launcher in front"
	}
	fmt.Fprintf(out, "smith %s runs %s.\n", s.Version, how)
	if s.Installed != "" {
		fmt.Fprintf(out, "Installed program: %s\n", s.Installed)
	}
	if s.Linked != "" {
		fmt.Fprintf(out, "Linked version:    %s\n", s.Linked)
	}
	if len(s.Entries) == 0 {
		fmt.Fprintln(out, "No smith entry of this installation was found.")
		return
	}
	for _, e := range s.Entries {
		state := "starts a launcher first"
		if e.Kind == Direct {
			state = "starts the program directly"
		}
		fmt.Fprintf(out, "  %s: %s\n", e.Path, state)
	}
	for _, e := range s.Entries {
		if e.Kind != Direct {
			fmt.Fprintln(out, "Run smith link to change that.")
			return
		}
	}
}
