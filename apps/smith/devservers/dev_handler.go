package devservers

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Command is `smith dev`.
func Command(version string) *cobra.Command {
	var all, dryRun bool
	cmd := &cobra.Command{
		Use:   "dev [projects...]",
		Short: "Start several projects together, such as a frontend and the API it calls",
		Long: `Start several projects together: what they depend on is built first, then
every start task runs at once, its output prefixed with the project's name, and
Ctrl+C stops them all. Name the projects, pass --all for every project with a
start task, or pick from a list. It is smith run start -p <projects> with room
for every one to run at once; --dry-run prints that command.`,
		Example: `  smith dev web api
  smith dev --all
  smith dev            # pick in the terminal`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			var startable []string
			for _, p := range s.Graph.Projects {
				if _, ok := p.Tasks["start"]; ok {
					startable = append(startable, p.Name)
				}
			}
			sort.Strings(startable)
			if len(startable) == 0 {
				return fmt.Errorf("no project has a start task")
			}
			chosen := args
			switch {
			case all:
				chosen = startable
			case len(chosen) == 0 && term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())):
				final, err := tea.NewProgram(newChooser(startable), tea.WithOutput(cmd.OutOrStdout())).Run()
				if err != nil {
					return err
				}
				if chosen = final.(chooser).Chosen(); len(chosen) == 0 {
					return nil
				}
			case len(chosen) == 0:
				return fmt.Errorf("name the projects to start, or pass --all (projects with a start task: %s)", strings.Join(startable, ", "))
			}
			for _, name := range chosen {
				if !contains(startable, name) {
					return fmt.Errorf("%s has no start task (projects with one: %s)", name, strings.Join(startable, ", "))
				}
			}
			// Room for every start task at once, plus the builds before them.
			parallel := len(chosen) + 4
			fmt.Fprintf(cmd.ErrOrStderr(), "→ smith run start -p %s --parallel %d\n", strings.Join(chosen, ","), parallel)
			if dryRun {
				return nil
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{"start"}, Projects: chosen, Parallel: parallel, Bail: true, Out: cmd.OutOrStdout()})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), summary.String())
			if !summary.OK() && ctx.Err() == nil {
				return taskrun.ErrTasksFailed
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "start every project that has a start task")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the equivalent smith run command and start nothing")
	return cmd
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
