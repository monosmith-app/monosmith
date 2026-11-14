package selfupdate

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/directlink"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/updatenotice"
)

// Command is smith update: update smith itself the way it was installed.
//
// Parameters: version is the running smith.
// Returns: the cobra command.
func Command(version string) *cobra.Command {
	var check, yes bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update smith itself to the newest release, the way it was installed",
		Long: `Ask the npm registry for the newest smith on your channel (stable, or next when you run a
prerelease) and update smith the way it was installed: npm, dotnet tool, pip, uv, pipx or the
install script. It shows the command first and runs it only after you choose.

This updates the smith program. To bring a workspace up to the smith you run, use smith upgrade.

On Windows a running smith cannot be replaced, so the update opens its own window and waits for
this command to exit; other smith programs of the same install (an editor's smith mcp, smith
serve) must be closed first, and smith update names them.`,
		Example: "  smith update\n  smith update --check\n  smith update --yes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), version, check, yes, os.Stdin, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "only say whether a newer smith exists and what would run; change nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "update without asking")
	return cmd
}

func run(ctx context.Context, version string, check, yes bool, in *os.File, out, errOut io.Writer) error {
	if ctx == nil {
		ctx = context.Background()
	}
	o := updatenotice.ForThisProcess(version, 0)
	u, err := updatenotice.Check(ctx, o, 10*time.Second)
	if err != nil {
		return problem.New("could not ask the registry for the newest smith: "+err.Error(), "check the network and run smith update again", "releases: "+updatenotice.ReleasesURL)
	}
	if !u.Available {
		fmt.Fprintf(out, "smith %s is the newest smith on the %s channel.\n", u.Current, u.Channel)
		return nil
	}
	fmt.Fprintf(out, "smith %s is out, you have %s. What's new: %s/tag/smith@%s\n", u.Latest, u.Current, updatenotice.ReleasesURL, u.Latest)
	if !u.Runnable() {
		return problem.New("smith cannot tell how it was installed, so it cannot update itself", "download "+u.Latest+" from "+updatenotice.ReleasesURL)
	}
	fmt.Fprintf(out, "It would run: %s\n", u.Command)
	if check {
		return nil
	}
	if blockers := updatenotice.BlockingProcesses(ctx, o.Exe, os.Getpid(), runtime.GOOS); len(blockers) > 0 {
		return problem.New("other smith programs use this install and Windows cannot replace it while they run: "+describe(blockers),
			"close them (the editor or assistant that started smith mcp, smith serve) and run smith update again")
	}
	if !yes {
		if !term.IsTerminal(int(in.Fd())) {
			return problem.New("smith update needs a choice and there is no terminal to ask", "run smith update --yes to update without asking", "run smith update --check to only look")
		}
		answer := choiceprompt.Ask("Update smith?", []choiceprompt.Option{
			{Label: "Update now", Value: "update", Hint: "runs " + u.Command},
			{Label: "Not now", Value: "no", Hint: "changes nothing; smith update asks again whenever you run it"},
		}, "update", bufio.NewReader(in), errOut)
		if answer != "update" {
			fmt.Fprintln(out, "Nothing changed.")
			return nil
		}
	}
	// A smith that runs from its link does not hold the installed file, so the update can run here.
	detach := runtime.GOOS == "windows" && !directlink.RunningFromLink(o.Home, directlink.ForThisProcess(version).Exe)
	detached, err := updatenotice.Run(ctx, u.Command, detach, os.Getpid(), out, errOut)
	switch {
	case errors.Is(err, updatenotice.ErrNotRunnable):
		return problem.New("smith cannot tell how it was installed", "download "+u.Latest+" from "+updatenotice.ReleasesURL)
	case err != nil:
		return problem.New(err.Error(), "run the command yourself: "+u.Command)
	case detached:
		fmt.Fprintln(out, "The update continues in a new window once this command exits.")
	default:
		fmt.Fprintf(out, "smith %s is installed. Run smith --version to see it.\n", u.Latest)
	}
	return nil
}

// describe words the blocking processes: "smith.exe mcp (PID 12), smith.exe serve (PID 40)".
func describe(blockers []updatenotice.Blocker) string {
	parts := make([]string, 0, len(blockers))
	for _, b := range blockers {
		parts = append(parts, fmt.Sprintf("%s (PID %d)", b.Command, b.PID))
	}
	return strings.Join(parts, ", ")
}
