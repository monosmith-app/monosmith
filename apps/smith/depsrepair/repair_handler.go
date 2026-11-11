package depsrepair

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// environment is what the command reads and opens. NewCommand gives the real
// one; a test gives one that opens no workspace and starts no plugin.
type environment struct {
	// interactive reports whether a person can answer on stdin.
	interactive func() bool
	// in is where the answers are read.
	in io.Reader
	// open loads the offer of the workspace for the targets, with the ports
	// that carry it out and a function that closes the workspace.
	open func(dir, version string, targets []string) (Offer, Ports, func(), error)
}

// NewCommand is `smith deps repair`.
//
// Parameters: version is smith's version, for the workspace session.
//
// Returns: the command.
func NewCommand(version string) *cobra.Command {
	return newCommand(version, environment{
		interactive: func() bool { return term.IsTerminal(int(os.Stdin.Fd())) },
		in:          os.Stdin,
		open: func(dir, version string, targets []string) (Offer, Ports, func(), error) {
			s, err := session.Open(session.Options{Dir: dir, Version: version, Log: os.Stderr})
			if err != nil {
				return Offer{}, Ports{}, func() {}, err
			}
			offer, err := Load(s, targets)
			if err != nil {
				s.Close()
				return Offer{}, Ports{}, func() {}, err
			}
			return offer, SessionPorts(s), func() { s.Close() }, nil
		},
	})
}

// kindNames is the --kind values, as one string for the messages.
func kindNames() string {
	var names []string
	for _, k := range protocol.RepairKinds() {
		names = append(names, string(k))
	}
	return strings.Join(names, "|")
}

// parseKind reads a --kind value or an answer: a kind's name or its number.
func parseKind(s string) (protocol.RepairKind, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	kinds := protocol.RepairKinds()
	if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= len(kinds) {
		return kinds[n-1], true
	}
	for _, k := range kinds {
		if string(k) == s {
			return k, true
		}
	}
	return "", false
}

func newCommand(version string, env environment) *cobra.Command {
	var kindFlag string
	var targets []string
	var dryRun, yes, asJSON bool
	cmd := &cobra.Command{
		Use:   "repair",
		Short: "Repair broken installed dependencies: prune, clean reinstall, rebuild the lockfile or clear the download cache",
		Long: `Repair the installed dependencies of the workspace when they are broken. Each
language plugin says how its tool does each kind, and you pick one kind each run:

  prune      remove the installed packages that no manifest declares
  reinstall  delete the installed dependencies, then install from the lockfile
             (the versions stay)
  lockfile   delete the installed dependencies and the lockfile, then install
             (the versions can change)
  cache      clear the tool's download cache, then install (the cache is shared
             by every workspace on this machine)

Without --kind, a terminal asks which kind. smith always prints the plan first:
what it deletes, what it runs and what it installs, tool by tool. Then it asks.
--dry-run prints the plan and changes nothing. Without a terminal, smith refuses
unless --kind and --yes are given.

smith deletes a link as a link: it never follows a symlink or a junction out of
the folder it deletes. It never commits: a rebuilt lockfile stays in the working
tree for you to review.`,
		Example: `  smith deps repair
  smith deps repair --kind reinstall
  smith deps repair --dry-run
  smith deps repair --kind lockfile -p web --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			var kind protocol.RepairKind
			if kindFlag != "" {
				k, ok := parseKind(kindFlag)
				if !ok {
					return problem.New(fmt.Sprintf("%q is not a repair kind", kindFlag), "use --kind "+kindNames())
				}
				kind = k
			}
			if !dryRun && (kind == "" || !yes) && !env.interactive() {
				return problem.New("refusing to repair dependencies without a terminal to ask in",
					"run `smith deps repair --kind <"+kindNames()+"> --yes`, or add --dry-run to see the plan")
			}
			if asJSON && !dryRun && (kind == "" || !yes) {
				return problem.New("--json cannot ask which repair to run or whether to go on",
					"add --dry-run to print the plans, or give --kind and --yes")
			}
			offer, ports, closeWorkspace, err := env.open(".", version, targets)
			defer closeWorkspace()
			if err != nil {
				return err
			}
			if len(offer.Tools) == 0 {
				remedies := append([]string(nil), offer.Notes...)
				remedies = append(remedies, "repair by hand as the language's page in docs/catalog says, or update the plugin when a new version repairs")
				return problem.New("no tool of this workspace offers a dependency repair", remedies...)
			}
			if kind != "" && !offer.Has(kind) {
				return problem.New(fmt.Sprintf("no tool of this workspace has the repair %q (%s)", kind, KindLabel(kind)),
					"run `smith deps repair --dry-run` to see the repairs each tool has")
			}
			if dryRun {
				kinds := offer.Kinds()
				if kind != "" {
					kinds = []protocol.RepairKind{kind}
				}
				if asJSON {
					return encode(out, reportOf(offer, kinds, ""))
				}
				fmt.Fprintln(out, Banner(kind))
				for i, k := range kinds {
					if i > 0 {
						fmt.Fprintln(out)
					}
					fmt.Fprint(out, PlanText(offer, k))
				}
				fmt.Fprintln(out, "\n● Dry run: nothing was deleted, run or installed.")
				return nil
			}
			reader := bufio.NewReader(env.in)
			if kind == "" {
				fmt.Fprintln(out, Banner(""))
				fmt.Fprintln(out, "Which repair? Each one says what it deletes. The lightest comes first.")
				for _, line := range ChoiceLines(offer) {
					fmt.Fprintln(out, line)
				}
				fmt.Fprint(out, "Repair (1-4, empty to cancel): ")
				answer, _ := reader.ReadString('\n')
				if strings.TrimSpace(answer) == "" {
					return errors.New("repair cancelled")
				}
				k, ok := parseKind(answer)
				if !ok || !offer.Has(k) {
					return problem.New(fmt.Sprintf("%q is not a repair this workspace has", strings.TrimSpace(answer)),
						"run `smith deps repair` again and answer with the number of a repair that a tool has")
				}
				kind = k
				fmt.Fprintln(out)
			}
			progress := out
			if asJSON {
				progress = cmd.ErrOrStderr()
			} else {
				fmt.Fprintln(out, Banner(kind))
			}
			fmt.Fprint(progress, PlanText(offer, kind))
			if !yes {
				fmt.Fprint(out, "\nRepair now? [y/N] ")
				answer, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(answer), "y") {
					return errors.New("repair cancelled")
				}
			}
			fmt.Fprintln(progress)
			if err := Apply(ctxOrBackground(cmd.Context()), offer, kind, ports, progress); err != nil {
				return &problem.Error{
					Problem: "the dependency repair stopped: " + err.Error(),
					Remedy:  []string{"fix what the tool reported, then run `smith deps repair --kind " + string(kind) + "` again"},
					Cause:   err,
				}
			}
			if asJSON {
				return encode(out, reportOf(offer, []protocol.RepairKind{kind}, kind))
			}
			fmt.Fprintln(out, Closing(offer, kind))
			return nil
		},
	}
	cmd.Flags().StringVar(&kindFlag, "kind", "", "the repair to run: "+kindNames()+" (without it a terminal asks)")
	commandflags.AddProjectFlag(cmd, &targets, "repair only these projects (a tool that installs the workspace as one still deletes its shared folders)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the plan of each repair and change nothing")
	commandflags.AddYesFlag(cmd, &yes, "do not ask before the repair (needs --kind without a terminal)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the plans, or the repair that ran, as JSON (needs --dry-run, or --kind with --yes)")
	return cmd
}

// encode writes v as indented JSON.
func encode(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ctxOrBackground is ctx, or the background context when cobra gave none.
func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
