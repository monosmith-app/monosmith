package convert

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/fullconversion"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// fullArgs are what a full-level conversion needs from the command.
type fullArgs struct {
	Root    string
	Git     *gitclient.Git
	Plan    *convertplan.Plan
	Version string
	Files   []string
	Strict  bool
	// Choices holds the flags' values; Given says which were given.
	Choices convertchoices.Choices
	Given   convertchoices.Given
	DryRun  bool
	JSON    bool
	Yes     bool
	// SkipInstall and SkipVerify leave out the single npm install and the
	// build, test, lint and doctor checks that follow it.
	SkipInstall bool
	SkipVerify  bool
}

// fullJSON is `smith convert --level full --json`: the conversion plan as it
// always is, and the full level's plan beside it.
type fullJSON struct {
	Plan *convertplan.Plan     `json:"plan"`
	Full fullconversion.Report `json:"full"`
}

// runFull is the full level: the conversion, then the removal of the old
// tooling the choices retire. The choices are settled and the whole plan is
// shown before anything changes; the removals are then asked one group at a
// time, after the conversion is applied.
func runFull(cmd *cobra.Command, a fullArgs) error {
	out := cmd.OutOrStdout()
	log := out
	if a.JSON {
		log = cmd.ErrOrStderr()
	}
	materialize := func(dir string) error {
		scratch := *a.Plan
		scratch.Root = dir
		_, err := Apply(&scratch, ApplyOptions{Version: a.Version, SkipInstall: true, Out: io.Discard})
		return err
	}
	view, err := fullconversion.MakeView(a.Root, a.Files, materialize)
	if err != nil {
		return fmt.Errorf("could not work out what the conversion leaves behind: %w", err)
	}
	defer view.Close()

	usage := legacyinventory.Detect(view.Tree)
	defaults := convertchoices.Defaults(usage)
	choices := withDefaults(a.Choices, a.Given, defaults)
	ask := interactive() && !a.Yes && !a.DryRun && !a.JSON
	reader := bufio.NewReader(os.Stdin)
	if ask {
		if choices, err = convertchoices.Ask(reader, log, convertchoices.Questions(usage, defaults), choices, a.Given); err != nil {
			return err
		}
	}
	report := fullconversion.MakeReport(view.Tree, choices, !a.SkipVerify)

	switch {
	case a.JSON:
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(fullJSON{Plan: a.Plan, Full: report}); err != nil {
			return err
		}
	case a.DryRun:
		if err := fullconversion.Describe(out, view.Dir, report, true); err != nil {
			return err
		}
	default:
		if err := fullconversion.Describe(log, view.Dir, report, false); err != nil {
			return err
		}
	}
	if a.DryRun {
		fmt.Fprintf(log, "\nNothing changed (dry run). To apply, with these choices:\n  → %s\n", report.Command)
		return nil
	}
	if !ask && !a.Yes {
		return problem.New("the full level changes the tooling and asks you first, which needs a terminal: nothing was changed",
			"smith convert --dry-run --level full shows every change", "to apply with these choices: "+report.Command)
	}

	if ask {
		fmt.Fprint(log, "\nThis converts the workspace, then asks about each group above before it deletes anything (git can undo it all). `smith convert --level full --dry-run` lists every change.\nConvert now? [y/N] ")
		answer, _ := reader.ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			return errors.New("convert cancelled: nothing changed")
		}
	}
	fmt.Fprintln(log)
	res, err := Apply(a.Plan, ApplyOptions{Version: a.Version, SkipInstall: true, KeepPipelines: choices.CI == convertchoices.Keep, Out: log})
	if err != nil {
		return err
	}
	if err := saveReport(a.Root, Report(a.Plan, ReportOptions{Verbose: true, Strict: a.Strict})); err != nil {
		fmt.Fprintf(log, "! could not save the report: %v\n", err)
	}
	staged := stageNewFiles(a.Git, a.Plan, res)

	smith := smithCommand()
	run, err := fullconversion.Run(cmd.Context(), fullconversion.RunOptions{
		Root: a.Root, Out: log, In: reader, Interactive: ask, Yes: a.Yes, Choices: choices,
		SkipFinish: a.SkipInstall, Verify: !a.SkipVerify, Smith: smith, Files: a.Git.Files,
	})
	if err != nil {
		fmt.Fprintf(log, "\n%v\n", err)
		fmt.Fprintln(log, "Everything above is in the working tree and was clean before, so git undoes it:\n  git restore --staged --worktree . && git clean -fd")
		return fmt.Errorf("the full conversion stopped: %w", err)
	}
	fmt.Fprint(log, Summary(a.Plan, res.LostSlots, staged))
	if len(run.Declined) > 0 {
		fmt.Fprintf(log, "\nLeft as they were, as you chose: %s\n", strings.Join(run.Declined, ", "))
	}
	return nil
}

// smithCommand is how the checks run smith: by name when the smith on the
// path is this one, else by the path of the running executable.
func smithCommand() string {
	self, err := os.Executable()
	if err != nil {
		return "smith"
	}
	if found, err := exec.LookPath("smith"); err == nil {
		a, errA := filepath.EvalSymlinks(found)
		b, errB := filepath.EvalSymlinks(self)
		if errA == nil && errB == nil && strings.EqualFold(a, b) {
			return "smith"
		}
	}
	return self
}

// withDefaults fills the choices the flags did not give from the defaults.
func withDefaults(flags convertchoices.Choices, given convertchoices.Given, defaults convertchoices.Choices) convertchoices.Choices {
	c := defaults
	c.Level = flags.Level
	for key, on := range map[string]bool{"lint": given.Lint, "test": given.Test, "build": given.Build, "ci": given.CI} {
		if on {
			c.Set(key, flags.Value(key))
		}
	}
	return c
}
