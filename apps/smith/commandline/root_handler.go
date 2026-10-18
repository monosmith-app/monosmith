package commandline

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/adoption"
	"github.com/monosmith-app/monosmith/apps/smith/devservers"
	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
	"github.com/monosmith-app/monosmith/apps/smith/npmaudit"
	"github.com/monosmith-app/monosmith/apps/smith/npmdepscheck"
	"github.com/monosmith-app/monosmith/apps/smith/npmpack"
	"github.com/monosmith-app/monosmith/apps/smith/npmprune"
	"github.com/monosmith-app/monosmith/apps/smith/pipelinetemplate"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/dependencies"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/explain"
	"github.com/monosmith-app/monosmith/apps/smith/graphview"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/pubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/slicecheck"
	"github.com/monosmith-app/monosmith/apps/smith/stackview"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/tui"
	"github.com/monosmith-app/monosmith/apps/smith/upgrade"
	"github.com/monosmith-app/monosmith/apps/smith/whymiss"
	"github.com/spf13/cobra"
)

// Root builds the smith command.
func Root(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "smith",
		Short: "A simple monorepo tool: add apps, libs and packages; build, test, release",
		Long: `smith runs a monorepo in Node, Go, Python, Rust, .NET, Ruby, PHP, C++, Dart, the JVM and Bicep (and anything a plugin adds):
it scaffolds projects, runs their tasks in dependency order from a cache,
generates CI pipelines, and releases packages from conventional commits.`,
		Example: `  smith                      open the interactive menu (in a terminal)
  smith new acme
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
	add("start", scaffold.NewCommand(version), scaffold.AddCommand(version), scaffold.RemoveCommand(version), convert.Command(version), configedit.Command(), upgrade.Command(version), adoption.Command(version))
	add("work", taskrun.Commands(version)...)
	add("work", devservers.Command(version))
	root.AddCommand(npmpack.Command(), npmprune.Command(), npmaudit.Command(), npmdepscheck.Command())
	add("work", dependencies.Commands(version)...)
	releaseCmd := release.Command(version)
	releaseCmd.AddCommand(pubrelease.Command(version))
	add("ship", releaseCmd, cipipeline.Command(version), pipelinetemplate.Command(version))
	add("inspect", graphview.Commands(version)...)
	add("inspect", clicontract.Commands(version)...)
	add("inspect", mcpserver.Command(version))
	add("inspect", doctor.Command(version), explain.Command(version), whymiss.Command(version), pluginhost.Command(version), stackview.Command(version), slicecheck.Command())
	return root
}

// Execute runs smith and returns the process exit code.
func Execute(version string, args []string) int {
	if len(args) == 0 && tui.Interactive() {
		chosen, err := tui.Choose(version, os.Stdout, clicontract.DescribeCommands(Root(version)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "smith: "+err.Error())
			return 1
		}
		if len(chosen) == 0 {
			return 0
		}
		args = chosen
	}
	root := Root(version)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if !errors.Is(err, taskrun.ErrTasksFailed) && !errors.Is(err, doctor.ErrUnhealthy) && !errors.Is(err, adoption.ErrBlocked) {
			fmt.Fprintln(os.Stderr, "smith: "+err.Error())
		}
		return 1
	}
	return 0
}
