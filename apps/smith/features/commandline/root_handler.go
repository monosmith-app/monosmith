package commandline

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/features/adoption"
	"github.com/monosmith-app/monosmith/apps/smith/features/agentdoc"
	"github.com/monosmith-app/monosmith/apps/smith/features/agentfiles"
	"github.com/monosmith-app/monosmith/apps/smith/features/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/features/completioncommand"
	"github.com/monosmith-app/monosmith/apps/smith/features/declarationspecifiers"
	"github.com/monosmith-app/monosmith/apps/smith/features/devservers"
	"github.com/monosmith-app/monosmith/apps/smith/features/directlink"
	"github.com/monosmith-app/monosmith/apps/smith/features/engineserver"
	"github.com/monosmith-app/monosmith/apps/smith/features/godepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/features/languageserver"
	"github.com/monosmith-app/monosmith/apps/smith/features/lockfilesync"
	"github.com/monosmith-app/monosmith/apps/smith/features/mcpserver"
	"github.com/monosmith-app/monosmith/apps/smith/features/npmaudit"
	"github.com/monosmith-app/monosmith/apps/smith/features/npmdepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/features/npmpack"
	"github.com/monosmith-app/monosmith/apps/smith/features/npmpolicy"
	"github.com/monosmith-app/monosmith/apps/smith/features/npmprune"
	"github.com/monosmith-app/monosmith/apps/smith/features/operation"
	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/features/pipelinetemplate"
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/features/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/features/projectimport"
	"github.com/monosmith-app/monosmith/apps/smith/features/pydepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/features/pypolicy"
	"github.com/monosmith-app/monosmith/apps/smith/features/rustpolicy"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/consolecolour"
	"golang.org/x/term"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/apptasks"
	"github.com/monosmith-app/monosmith/apps/smith/features/bootstrap"
	"github.com/monosmith-app/monosmith/apps/smith/features/boundaries"
	"github.com/monosmith-app/monosmith/apps/smith/features/cachecommand"
	"github.com/monosmith-app/monosmith/apps/smith/features/choiceview"
	"github.com/monosmith-app/monosmith/apps/smith/features/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/features/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/features/composeview"
	"github.com/monosmith-app/monosmith/apps/smith/features/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/features/convert"
	"github.com/monosmith-app/monosmith/apps/smith/features/copymeasure"
	"github.com/monosmith-app/monosmith/apps/smith/features/dependencies"
	"github.com/monosmith-app/monosmith/apps/smith/features/derive"
	"github.com/monosmith-app/monosmith/apps/smith/features/divergences"
	"github.com/monosmith-app/monosmith/apps/smith/features/dockerdescription"
	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/eject"
	"github.com/monosmith-app/monosmith/apps/smith/features/environments"
	"github.com/monosmith-app/monosmith/apps/smith/features/explain"
	"github.com/monosmith-app/monosmith/apps/smith/features/graphview"
	"github.com/monosmith-app/monosmith/apps/smith/features/kindchange"
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginusage"
	"github.com/monosmith-app/monosmith/apps/smith/features/prompt"
	"github.com/monosmith-app/monosmith/apps/smith/features/pubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/features/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/features/recipewizard"
	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/releaseassets"
	"github.com/monosmith-app/monosmith/apps/smith/features/removeproject"
	"github.com/monosmith-app/monosmith/apps/smith/features/report"
	"github.com/monosmith-app/monosmith/apps/smith/features/sandbox"
	"github.com/monosmith-app/monosmith/apps/smith/features/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/features/selfupdate"
	"github.com/monosmith-app/monosmith/apps/smith/features/slicecheck"
	"github.com/monosmith-app/monosmith/apps/smith/features/stackview"
	"github.com/monosmith-app/monosmith/apps/smith/features/swapchoice"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/features/tauriadd"
	"github.com/monosmith-app/monosmith/apps/smith/features/tauriprepare"
	"github.com/monosmith-app/monosmith/apps/smith/features/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui"
	"github.com/monosmith-app/monosmith/apps/smith/features/upgrade"
	"github.com/monosmith-app/monosmith/apps/smith/features/why"
	"github.com/monosmith-app/monosmith/apps/smith/features/whymiss"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
	"github.com/spf13/cobra"
)

// withWizard adds `smith recipe new` to `smith recipe`: the wizard needs the
// recipe package, so it cannot be added from inside it.
func withWizard(cmd *cobra.Command, version string) *cobra.Command {
	cmd.AddCommand(recipewizard.NewCommand(version))
	return cmd
}

// applyConfirmSetting reads [workspace] confirm from the workspace the command
// runs in, so every plan presents itself the way the person set. A missing or
// broken smith.toml leaves the default (every plan asks).
func applyConfirmSetting() {
	root, err := workspace.FindRoot(".")
	if err != nil {
		return
	}
	if cfg, err := workspace.Load(root); err == nil {
		operationplan.DefaultConfirm = cfg.Workspace.Confirm
	}
}

// Root builds the smith command.
func Root(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "smith",
		Short: "A simple monorepo tool: add apps, libs and packages; build, test, release",
		Long: `smith runs a monorepo in Node, Go, Python, Rust, .NET, Ruby, PHP, C++, Dart, the JVM, Bicep, Lua, AppleScript, VBA, Swift and Objective-C (and anything a plugin adds):
it scaffolds projects, runs their tasks in dependency order from a cache,
generates CI pipelines, and releases packages from conventional commits.

Most people start with: new (an empty folder) or adopt (a repository you
already have), then build and doctor.`,
		Example: `  smith                      open the interactive shell (in a terminal: ? lists the keys, : the commands)
  smith new acme             a new, empty workspace
  smith adopt --dry-run      what bringing the repository you are in under smith involves
  smith add app web --lang node
  smith build`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		// [workspace] confirm says which plans ask; every plan reads it from here.
		PersistentPreRun: func(*cobra.Command, []string) { applyConfirmSetting() },
	}
	root.AddGroup(&cobra.Group{ID: "start", Title: "Create:"}, &cobra.Group{ID: "work", Title: "Work:"}, &cobra.Group{ID: "ship", Title: "Ship:"}, &cobra.Group{ID: "inspect", Title: "Inspect:"})
	add := func(group string, cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.GroupID = group
			root.AddCommand(c)
		}
	}
	add("start", scaffold.NewCommand(version), scaffold.AddCommand(version), removeproject.RemoveCommand(version), swapchoice.Command(version), environments.Command(version), convert.Command(version), configedit.Command(), architecture.Command(version), upgrade.Command(version), adoption.Command(version), projectimport.Command(version), bootstrap.Command(), sandbox.Command())
	add("work", taskrun.Commands(version)...)
	add("work", selfupdate.Command(version), directlink.Command(version), toolinstall.Command(version), devservers.Command(version), derive.Command(version), operation.UndoCommand(), operation.CommitCommand(), eject.Command(version), kindchange.Command(version), tauriadd.Command(version))
	root.AddCommand(apptasks.FuncCommand(), apptasks.ElectronCommand(), declarationspecifiers.Command(), npmpack.Command(), npmprune.Command(), tauriprepare.Command(version), lockfilesync.Command(), npmaudit.Command(), npmdepscheck.Command(), pydepscheck.Command(), godepscheck.Command(), npmpolicy.Command(), pypolicy.Command(), rustpolicy.Command(), completioncommand.Command(plugintrust.Home()))
	add("work", dependencies.Commands(version)...)
	releaseCmd := release.Command(version)
	releaseCmd.AddCommand(pubrelease.Command(version), releaseassets.Command(version, cipipeline.AssetsPrepare))
	add("work", cipipeline.VerifyCommand(version))
	add("ship", releaseCmd, cipipeline.Command(version), pipelinetemplate.Command(version), dockerdescription.Command(version))
	add("inspect", graphview.Commands(version)...)
	add("inspect", clicontract.Commands(version)...)
	mcpCmd := mcpserver.Command(version)
	mcpCmd.AddCommand(agentfiles.McpInstallCommand())
	agentCmd := agentfiles.Command()
	agentCmd.AddCommand(agentdoc.FileCommand(version))
	add("inspect", mcpCmd, agentCmd, engineserver.Command(version), languageserver.Command(version))
	add("inspect", divergences.Command(version), copymeasure.Command(version), doctor.Command(version), prompt.Command(version), explain.Command(version), why.Command(version), whymiss.Command(version), report.Command(version), cachecommand.Command(), pluginCommand(version), withWizard(recipe.Command(version), version), stackview.Command(version), choiceview.Command(version), composeview.Command(version), slicecheck.Command(), boundaries.Command(version))
	return root
}

// canAsk is whether a person is at the terminal to be asked to trust an
// external plugin: stdin and stderr are terminals, this is not CI, and the
// command is not one that serves other programs (smith serve, smith mcp, smith lsp) or
// owns the screen (smith tui), which never ask, or a shell prompt calls on every
// line (smith prompt), which must answer at once and never wait for a person.
func canAsk(root *cobra.Command, args []string) bool {
	if plugintrust.InCI(os.Getenv) || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}
	cmd, _, err := root.Find(args)
	if err != nil || cmd == nil {
		return true
	}
	for cmd.Parent() != nil && cmd.Parent().Parent() != nil {
		cmd = cmd.Parent()
	}
	switch cmd.Name() {
	case "serve", "mcp", "lsp", "tui", "prompt", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		// The last two are what the completion script of a shell calls on every Tab.
		return false
	}
	return true
}

// Execute runs smith and returns the process exit code. After the first
// interactive run (the interface, or smith doctor in a terminal) it offers,
// once, to set up tab completion; see offerCompletion.
func Execute(version string, args []string) int {
	// The colour smith and its tools print renders in a classic Windows console too.
	consolecolour.Enable()
	// The first run of a version points the smith command at the program itself, and says so at the end.
	done, linked := linkFirstRun(args, version)
	if len(args) > 0 || !tui.Interactive() {
		defer done()
		code := run(version, args)
		if offersCompletionAfter(args) {
			offerCompletion(version)
		}
		return code
	}
	// The interface says it itself and offers to restart: the line at its end would repeat it.
	code := openInterface(version, linked)
	offerCompletion(version)
	return code
}

// openInterface runs the interactive shell and the commands it hands back.
func openInterface(version string, linked bool) int {
	// A bare smith at a terminal opens the interface. A command it hands back
	// runs here, and when that command made a workspace the interface opens
	// again inside it: nobody should have to cd and start smith a second time.
	var resume *tui.Resume
	for {
		chosen, err := tui.Choose(version, os.Stdout, clicontract.DescribeCommands(Root(version)), resume, linked)
		if err != nil {
			fmt.Fprintln(os.Stderr, "smith: "+err.Error())
			return 1
		}
		resume, linked = nil, false
		if chosen.Restart {
			return restartInterface()
		}
		argv := chosen.Argv
		if len(argv) == 0 {
			return 0
		}
		code, message := runCommand(version, argv)
		if code != 0 && len(argv) > 1 && argv[0] == "new" && chosen.Answers != nil {
			// A failed `new` does not end the session: the form opens again with
			// the answers kept and the reason under the first question.
			resume = &tui.Resume{Form: chosen.Form, Answers: chosen.Answers, Problem: strings.TrimPrefix(message, "smith: ")}
			continue
		}
		if message != "" {
			fmt.Fprintln(os.Stderr, message)
		}
		if code != 0 || len(argv) < 2 || argv[0] != "new" {
			return code
		}
		if err := os.Chdir(argv[1]); err != nil {
			return 0
		}
	}
}

// run executes one smith command line, prints what went wrong and returns the
// exit code.
func run(version string, args []string) int {
	code, message := runCommand(version, args)
	if message != "" {
		fmt.Fprintln(os.Stderr, message)
	}
	return code
}

// runCommand executes one smith command line and returns the exit code and
// the line that says why it failed ("" when nothing is left to say).
func runCommand(version string, args []string) (int, string) {
	root := Root(version)
	root.SetArgs(args)
	if canAsk(root, args) {
		pluginadmit.SetPrompter(&pluginadmit.Prompter{Out: os.Stderr, In: bufio.NewReader(os.Stdin)})
	}
	announce := startUpdateNotice(root, args, version)
	defer announce()
	if err := root.Execute(); err != nil {
		var p *problem.Error
		message, newer := newerSmithMessage(err, version, wantsVerbose(args))
		switch {
		case errors.Is(err, release.ErrPropagationPending), errors.Is(err, convert.ErrFindings):
			// The report said what to run; code 2 is a warning, not a failure.
			return 2, ""
		case newer:
			return 1, "smith: " + message
		case errors.As(err, &p):
			return 1, p.Error()
		case !errors.Is(err, taskrun.ErrTasksFailed) && !errors.Is(err, doctor.ErrUnhealthy) && !errors.Is(err, adoption.ErrBlocked) && !errors.Is(err, recipe.ErrInvalid) && !errors.Is(err, release.ErrReported) && !errors.Is(err, toolinstall.ErrMissing):
			return 1, "smith: " + err.Error()
		}
		return 1, ""
	}
	return 0, ""
}

// pluginCommand is `smith plugin` with the check that stops switching off a
// plugin some project uses, unless --force says it is meant.
func pluginCommand(version string) *cobra.Command {
	cmd := pluginhost.Command(version)
	pluginusage.GuardDisable(cmd, version)
	return cmd
}
