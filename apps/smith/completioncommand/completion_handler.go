package completioncommand

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Dependencies are what the command reads and writes. Command uses the real
// machine; the tests use a temporary folder.
type Dependencies struct {
	// Env is the machine.
	Env shellenv.Environment
	// Memory keeps the answer to the offer.
	Memory Memory
	// Interactive is true when a person can answer a question.
	Interactive bool
}

// Command is `smith completion`: it installs, removes and reports the tab
// completion of smith, and prints the scripts for the shells (`smith completion
// bash`, as cobra's own command did).
//
// Parameters: home is the smith home that keeps the answer to the offer.
func Command(home string) *cobra.Command {
	return NewCommand(Dependencies{
		Env: shellenv.OSEnvironment{}, Memory: PreferencesMemory{Home: home},
		Interactive: term.IsTerminal(int(os.Stdin.Fd())),
	})
}

// NewCommand is Command with its dependencies given.
func NewCommand(d Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Set up tab completion in your shell",
		Long: `completion adds tab completion for smith to your shell, when you ask for it.
Smith never sets it up on its own: install shows the files and the lines it
writes, and asks first. Uninstall removes exactly what install added.`,
		Example: `  smith completion install
  smith completion install --shell powershell
  smith completion status
  smith completion uninstall`,
	}
	cmd.AddCommand(installCommand(d), uninstallCommand(d), statusCommand(d))
	for _, s := range shellenv.Shells {
		cmd.AddCommand(scriptCommand(s))
	}
	return cmd
}

// scriptCommand is `smith completion <shell>`: it prints the script.
func scriptCommand(shell shellenv.Shell) *cobra.Command {
	var plain bool
	c := &cobra.Command{
		Use:   string(shell),
		Short: "Print the completion script for " + shell.Title(),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			script, err := shellcompletion.ScriptFor(cmd.Root(), shell, plain)
			if err != nil {
				return err
			}
			_, err = io.WriteString(cmd.OutOrStdout(), script)
			return err
		},
	}
	c.Flags().BoolVar(&plain, "no-descriptions", false, "leave the descriptions out of the script")
	return c
}

// pickShell is the shell the person named, or the one found.
func pickShell(env shellenv.Environment, named string) (shellenv.Shell, error) {
	if named != "" {
		if s, ok := shellenv.ParseShell(named); ok {
			return s, nil
		}
		return "", problem.New(words.Say("completion.error.unknown_shell", "shell", named), words.Say("completion.error.pass_shell"))
	}
	if s, ok := shellenv.DetectShell(env); ok {
		return s, nil
	}
	return "", problem.New(words.Say("completion.error.no_shell"), words.Say("completion.error.pass_shell"))
}

// installCommand is `smith completion install`.
func installCommand(d Dependencies) *cobra.Command {
	var shellName string
	var yes, dry bool
	c := &cobra.Command{
		Use:   "install",
		Short: "Install tab completion for your shell",
		Long: `install writes the completion script to the standard user place of the shell,
between marker lines, after it shows the files and the lines and you say yes.
A second run changes nothing. When a package of smith (Homebrew, for example)
already ships the file, install says so and writes nothing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInstall(cmd, d, shellName, yes, dry)
		},
	}
	c.Flags().StringVar(&shellName, "shell", "", "bash, zsh, fish or powershell (default: the shell you are in)")
	commandflags.AddYesFlag(c, &yes, "write without asking (for scripts and CI)")
	c.Flags().BoolVar(&dry, "dry-run", false, "show what would be written and write nothing")
	return c
}

// runInstall plans, asks and writes.
func runInstall(cmd *cobra.Command, d Dependencies, shellName string, yes, dry bool) error {
	shell, err := pickShell(d.Env, shellName)
	if err != nil {
		return err
	}
	script, err := shellcompletion.ScriptFor(cmd.Root(), shell, false)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	plan := shellcompletion.PlanFor(d.Env, shell, script)
	if plan.PackagedAt != "" {
		fmt.Fprintln(out, words.Say("completion.packaged", "shell", shell.Title(), "path", plan.PackagedAt))
		return nil
	}
	if len(plan.Targets) == 0 {
		return problem.New(words.Say("completion.error.no_home"), words.Say("completion.error.set_home"))
	}
	if t, blocked := plan.Blocked(); blocked {
		return problem.New(words.Say("completion.error.foreign", "path", t.Path), words.Say("completion.error.move_away"))
	}
	if plan.Installed() {
		fmt.Fprintln(out, words.Say("completion.already", "shell", shell.Title()))
		return d.Memory.Save(Answer{Choice: ChoiceInstalled})
	}
	fmt.Fprintln(out, shellcompletion.RenderPlan(plan))
	if dry {
		fmt.Fprintln(out, words.Say("tui.capture.dry_run"))
		return nil
	}
	if !yes {
		if !d.Interactive {
			return problem.New(words.Say("completion.error.cannot_ask"), words.Say("completion.error.pass_yes"))
		}
		if Ask(Terminal{In: bufio.NewReader(cmd.InOrStdin()), Out: out}, plan, false) != AnswerYes {
			fmt.Fprintln(out, words.Say("completion.nothing_written"))
			return nil
		}
	}
	changes, err := shellcompletion.Install(d.Env, plan)
	shellcompletion.PrintChanges(out, shell, changes)
	if err != nil {
		return err
	}
	return d.Memory.Save(Answer{Choice: ChoiceInstalled})
}

// uninstallCommand is `smith completion uninstall`.
func uninstallCommand(d Dependencies) *cobra.Command {
	var shellName string
	var yes bool
	c := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the tab completion that install added",
		Long: `uninstall removes the marked block that install wrote, and the file when
nothing else is left in it. Without --shell it looks at every shell.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUninstall(cmd, d, shellName, yes)
		},
	}
	c.Flags().StringVar(&shellName, "shell", "", "bash, zsh, fish or powershell (default: every shell that has it)")
	commandflags.AddYesFlag(c, &yes, "remove without asking")
	return c
}

// runUninstall lists what holds a block, asks and removes it.
func runUninstall(cmd *cobra.Command, d Dependencies, shellName string, yes bool) error {
	shells := shellenv.Shells
	if shellName != "" {
		s, err := pickShell(d.Env, shellName)
		if err != nil {
			return err
		}
		shells = []shellenv.Shell{s}
	}
	var found []string
	for _, s := range shells {
		for _, t := range shellcompletion.Targets(d.Env, s, "") {
			if raw, err := d.Env.ReadFile(t.Path); err == nil && shellcompletion.HasBlock(string(raw)) {
				found = append(found, t.Path)
			}
		}
	}
	out := cmd.OutOrStdout()
	if len(found) == 0 {
		fmt.Fprintln(out, words.Say("completion.nothing_to_remove"))
		return d.Memory.Save(Answer{})
	}
	fmt.Fprintln(out, words.Say("completion.remove.intro", "files", "\n  "+strings.Join(found, "\n  ")))
	if !yes {
		if !d.Interactive {
			return problem.New(words.Say("completion.error.cannot_ask"), words.Say("completion.error.pass_yes"))
		}
		options := []choiceprompt.Option{
			{Label: words.Say("completion.ask.yes"), Value: AnswerYes, Hint: words.Say("completion.remove.yes.hint")},
			{Label: words.Say("completion.ask.no"), Value: AnswerNo, Hint: words.Say("completion.ask.no.hint")},
		}
		if choiceprompt.Ask(words.Say("completion.remove.question"), options, AnswerNo, bufio.NewReader(cmd.InOrStdin()), out) != AnswerYes {
			fmt.Fprintln(out, words.Say("completion.nothing_removed"))
			return nil
		}
	}
	for _, s := range shells {
		changes, err := shellcompletion.Uninstall(d.Env, s)
		shellcompletion.PrintChanges(out, s, changes)
		if err != nil {
			return err
		}
	}
	return d.Memory.Save(Answer{})
}

// StatusReport is what `smith completion status --json` prints. An editor
// reads it to decide whether to offer the setup.
type StatusReport struct {
	// Shell is the shell found, "" when none was.
	Shell string `json:"shell"`
	// Wanted is true when nothing is installed for that shell, no package ships
	// the file, and the person did not say no: an offer is in order.
	Wanted bool `json:"wanted"`
	// Answer is what the person answered to the offer: "no", "later",
	// "installed" or "" for not asked yet.
	Answer string `json:"answer"`
	// Stale is true when a script that smith wrote for that shell comes from
	// another version of smith: `smith completion install` rewrites it.
	Stale bool `json:"stale"`
}

// statusCommand is `smith completion status`.
func statusCommand(d Dependencies) *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "status",
		Short: "Show where tab completion is installed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !asJSON {
				printStatus(cmd, d)
				return nil
			}
			shell, _ := shellenv.DetectShell(d.Env)
			answer := d.Memory.Load()
			report := StatusReport{Shell: string(shell), Wanted: HomeOfferWanted(d.Env, answer), Answer: answer.Choice}
			if shell != "" {
				report.Stale = isStale(cmd, d, shell)
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "print the state as JSON, for editors")
	return c
}

// isStale is true when a file that smith wrote for the shell holds the script
// of another smith: its block is there, and it differs from the script of this
// smith.
func isStale(cmd *cobra.Command, d Dependencies, shell shellenv.Shell) bool {
	script, err := shellcompletion.ScriptFor(cmd.Root(), shell, false)
	if err != nil {
		return false
	}
	for _, t := range shellcompletion.Targets(d.Env, shell, script) {
		if shellcompletion.StateOf(d.Env, t) == shellcompletion.Outdated {
			return true
		}
	}
	return false
}

// printStatus writes one line for each shell: installed (and where, and
// whether the script is stale), shipped by the package, or not installed. The
// shell found is marked.
func printStatus(cmd *cobra.Command, d Dependencies) {
	out := cmd.OutOrStdout()
	detected, _ := shellenv.DetectShell(d.Env)
	for _, s := range shellenv.Shells {
		state := words.Say("completion.status.none")
		var paths []string
		for _, t := range shellcompletion.Targets(d.Env, s, "") {
			if raw, err := d.Env.ReadFile(t.Path); err == nil && shellcompletion.HasBlock(string(raw)) {
				paths = append(paths, t.Path)
			}
		}
		if len(paths) > 0 {
			state = words.Say("completion.status.installed", "paths", strings.Join(paths, ", "))
			if isStale(cmd, d, s) {
				state += " " + words.Say("completion.status.stale")
			}
		} else if p := shellcompletion.PackagedPath(d.Env, s); p != "" {
			state = words.Say("completion.status.packaged", "path", p)
		}
		mark := ""
		if s == detected {
			mark = words.Say("completion.status.yours")
		}
		fmt.Fprintf(out, "%-11s %s%s\n", s.Title(), state, mark)
	}
	fmt.Fprintln(out, words.Say("completion.status.answer."+answerName(d.Memory.Load().Choice)))
}

// answerName is the key part of a remembered answer; "" is "none".
func answerName(choice string) string {
	if choice == "" {
		return "none"
	}
	return choice
}
