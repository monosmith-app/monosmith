package commandline

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/adoption"
	"github.com/monosmith-app/monosmith/apps/smith/agentfiles"
	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/declarationspecifiers"
	"github.com/monosmith-app/monosmith/apps/smith/devservers"
	"github.com/monosmith-app/monosmith/apps/smith/engineserver"
	"github.com/monosmith-app/monosmith/apps/smith/godepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/languageserver"
	"github.com/monosmith-app/monosmith/apps/smith/lockfilesync"
	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
	"github.com/monosmith-app/monosmith/apps/smith/npmaudit"
	"github.com/monosmith-app/monosmith/apps/smith/npmdepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/npmpack"
	"github.com/monosmith-app/monosmith/apps/smith/npmpolicy"
	"github.com/monosmith-app/monosmith/apps/smith/npmprune"
	"github.com/monosmith-app/monosmith/apps/smith/pipelinetemplate"
	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/pydepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/pypolicy"
	"github.com/monosmith-app/monosmith/apps/smith/rustpolicy"
	"golang.org/x/term"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/boundaries"
	"github.com/monosmith-app/monosmith/apps/smith/cachecommand"
	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/dependencies"
	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/dockerdescription"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/eject"
	"github.com/monosmith-app/monosmith/apps/smith/explain"
	"github.com/monosmith-app/monosmith/apps/smith/graphview"
	"github.com/monosmith-app/monosmith/apps/smith/kindchange"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/pubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/releaseassets"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/slicecheck"
	"github.com/monosmith-app/monosmith/apps/smith/stackview"
	"github.com/monosmith-app/monosmith/apps/smith/choiceview"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/tauriadd"
	"github.com/monosmith-app/monosmith/apps/smith/tauriprepare"
	"github.com/monosmith-app/monosmith/apps/smith/tui"
	"github.com/monosmith-app/monosmith/apps/smith/upgrade"
	"github.com/monosmith-app/monosmith/apps/smith/whymiss"
	"github.com/spf13/cobra"
)

// withWizard adds `smith recipe new` to `smith recipe`: the wizard needs the
// recipe package, so it cannot be added from inside it.
func withWizard(cmd *cobra.Command, version string) *cobra.Command {
	cmd.AddCommand(recipewizard.NewCommand(version))
	return cmd
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
	}
	root.AddGroup(&cobra.Group{ID: "start", Title: "Create:"}, &cobra.Group{ID: "work", Title: "Work:"}, &cobra.Group{ID: "ship", Title: "Ship:"}, &cobra.Group{ID: "inspect", Title: "Inspect:"})
	add := func(group string, cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.GroupID = group
			root.AddCommand(c)
		}
	}
	add("start", scaffold.NewCommand(version), scaffold.AddCommand(version), scaffold.RemoveCommand(version), convert.Command(version), configedit.Command(), architecture.Command(version), upgrade.Command(version), adoption.Command(version))
	add("work", taskrun.Commands(version)...)
	add("work", devservers.Command(version), derive.Command(version), eject.Command(version), kindchange.Command(version), tauriadd.Command(version))
	root.AddCommand(declarationspecifiers.Command(), npmpack.Command(), npmprune.Command(), tauriprepare.Command(version), lockfilesync.Command(), npmaudit.Command(), npmdepscheck.Command(), pydepscheck.Command(), godepscheck.Command(), npmpolicy.Command(), pypolicy.Command(), rustpolicy.Command())
	add("work", dependencies.Commands(version)...)
	releaseCmd := release.Command(version)
	releaseCmd.AddCommand(pubrelease.Command(version), releaseassets.Command(version))
	add("ship", releaseCmd, cipipeline.Command(version), pipelinetemplate.Command(version), dockerdescription.Command(version))
	add("inspect", graphview.Commands(version)...)
	add("inspect", clicontract.Commands(version)...)
	mcpCmd := mcpserver.Command(version)
	mcpCmd.AddCommand(agentfiles.McpInstallCommand())
	add("inspect", mcpCmd, agentfiles.Command(), engineserver.Command(version), languageserver.Command(version))
	add("inspect", doctor.Command(version), explain.Command(version), whymiss.Command(version), cachecommand.Command(), pluginhost.Command(version), withWizard(recipe.Command(version), version), stackview.Command(version), choiceview.Command(version), slicecheck.Command(), boundaries.Command(version))
	return root
}

// canAsk is whether a person is at the terminal to be asked to trust an
// external plugin: stdin and stderr are terminals, this is not CI, and the
// command is not one that serves other programs (smith serve, smith mcp, smith lsp) or
// owns the screen (smith tui), which never ask.
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
	case "serve", "mcp", "lsp", "tui":
		return false
	}
	return true
}

// Execute runs smith and returns the process exit code.
func Execute(version string, args []string) int {
	if len(args) > 0 || !tui.Interactive() {
		return run(version, args)
	}
	// A bare smith at a terminal opens the interface. A command it hands back
	// runs here, and when that command made a workspace the interface opens
	// again inside it: nobody should have to cd and start smith a second time.
	for {
		chosen, err := tui.Choose(version, os.Stdout, clicontract.DescribeCommands(Root(version)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "smith: "+err.Error())
			return 1
		}
		if len(chosen) == 0 {
			return 0
		}
		if code := run(version, chosen); code != 0 || len(chosen) < 2 || chosen[0] != "new" {
			return code
		}
		if err := os.Chdir(chosen[1]); err != nil {
			return 0
		}
	}
}

// run executes one smith command line and returns the exit code.
func run(version string, args []string) int {
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
		case newer:
			fmt.Fprintln(os.Stderr, "smith: "+message)
		case errors.As(err, &p):
			fmt.Fprintln(os.Stderr, p.Error())
		case !errors.Is(err, taskrun.ErrTasksFailed) && !errors.Is(err, doctor.ErrUnhealthy) && !errors.Is(err, adoption.ErrBlocked) && !errors.Is(err, recipe.ErrInvalid):
			fmt.Fprintln(os.Stderr, "smith: "+err.Error())
		}
		return 1
	}
	return 0
}
