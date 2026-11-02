package environments

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
	"github.com/spf13/cobra"
)

// Command is `smith environments`: list the environments of the web apps, add
// one or remove one.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environments",
		Short: "List, add or remove the environments a web app is built for",
		Long: `A web app (React, Vue, Svelte or a plain page built with Vite) is built once per
environment, because Vite writes its VITE_ values into the bundle when it builds. Each
environment has its own .env.<name> file, a build script and an output folder. The default
environment is the task build; the others are build:<name>.

  smith environments list [project]
  smith environments add <project> <name>...
  smith environments remove <project> <name>...

add and remove show a plan first: the files they create, the scripts they change and the
line in smith.toml. remove leaves the .env.<name> file, which holds your values. The default
environment cannot be removed.`,
		Example: "  smith environments list\n  smith environments add web qa --dry-run\n  smith environments remove web dev --yes",
	}
	cmd.AddCommand(listCommand(version), changeCommand(version, "add"), changeCommand(version, "remove"))
	return cmd
}

// openApp opens the session and finds the project.
func openApp(cmd *cobra.Command, version, name string) (*session.Session, *projectgraph.Project, error) {
	s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
	if err != nil {
		return nil, nil, err
	}
	p := s.Graph.Get(name)
	if p == nil {
		defer s.Close()
		return nil, nil, fmt.Errorf("no project named %q (projects: %s).%s", name, strings.Join(s.Graph.Names(), ", "), nearmatch.Suggestion(name, s.Graph.Names()))
	}
	return s, p, nil
}

func listCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "list [project]",
		Short: "List the environments of the web apps and the task that builds each",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			out := cmd.OutOrStdout()
			found := false
			for _, p := range s.Graph.Projects {
				if len(args) == 1 && p.Name != args[0] {
					continue
				}
				settings, ok := webenv.FromMap(p.Settings[webenv.PluginName])
				if !ok {
					continue
				}
				found = true
				fmt.Fprintf(out, "%s (env files: %s, builds: %s)\n", p.Name, settings.EnvFiles, settings.Output)
				for _, env := range settings.Names {
					fmt.Fprintf(out, "  %-8s smith run %s%s\n", env, webenv.TaskName(settings, env), defaultMark(settings, env))
				}
			}
			if !found {
				fmt.Fprintln(out, "No web app here is built per environment. smith add app <name> --with react --environments lcl,dev,uat,prd makes one.")
			}
			return nil
		},
	}
}

func defaultMark(s webenv.Settings, env string) string {
	if s.Default == env {
		return "   (default: CI and releases build this one)"
	}
	return ""
}

func changeCommand(version, verb string) *cobra.Command {
	short := "Add environments to a web app"
	if verb == "remove" {
		short = "Remove environments from a web app (its .env files stay)"
	}
	cmd := &cobra.Command{
		Use:   verb + " <project> <name>...",
		Short: short,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, p, err := openApp(cmd, version, args[0])
			if err != nil {
				return err
			}
			defer s.Close()
			names := args[1:]
			if err := webenv.ValidateNames(names); err != nil {
				return err
			}
			var change Change
			if verb == "add" {
				change, err = PlanChange(p, names, nil)
			} else {
				change, err = PlanChange(p, nil, names)
			}
			if err != nil {
				return err
			}
			yes, _ := cmd.Flags().GetBool("yes")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			out := cmd.OutOrStdout()
			proceed, err := operationplan.Present(out, cmd.InOrStdin(), change.Plan(), operationplan.Options{Yes: yes, DryRun: dryRun, Ask: true, Env: operationplan.DetectEnvironment()})
			if err != nil || !proceed {
				return err
			}
			written, err := Apply(s.Root, change)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Wrote %d file(s):\n  %s\n", len(written), strings.Join(written, "\n  "))
			for _, kept := range change.Kept {
				fmt.Fprintf(out, "Left %s alone: it holds your values. Delete it yourself when you do not need it.\n", kept)
			}
			fmt.Fprintln(out, "Build them all: smith run "+strings.Join(buildTasks(change.After), " ")+" -p "+change.Project)
			return nil
		},
	}
	return operation.Wrap(cmd, operation.Spec{AddYes: true, AddDryRun: true, Describe: func(_ *cobra.Command, args []string) string {
		return "environments " + verb + " " + strings.Join(args, " ")
	}})
}

// buildTasks are the tasks that build every environment, the default first.
func buildTasks(s webenv.Settings) []string {
	var out []string
	for _, env := range s.Names {
		out = append(out, webenv.TaskName(s, env))
	}
	return out
}
