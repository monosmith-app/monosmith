package graphview

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Commands returns `smith graph`, `smith projects` and `smith affected`.
func Commands(version string) []*cobra.Command {
	var format, file, graphBase, diffFrom string
	var focus []string
	var onlyAffected bool
	graph := &cobra.Command{
		Use:   "graph",
		Short: "Show how projects depend on each other",
		Long: `Show the project graph as a tree (text), JSON, Graphviz (dot), a static page
(html) or an interactive browser (tui). Every format shows the same projects:
--focus keeps the named projects and everything they depend on or that depends
on them; --affected keeps the projects affected by changes since --base and
marks the ones that changed.`,
		Example: `  smith graph                       tree in the terminal
  smith graph --format tui          browse it
  smith graph --focus web           web, what it uses and what uses it
  smith graph --affected --format html -o affected.html
  smith graph --format dot | dot -Tsvg > graph.svg`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			for _, c := range s.Graph.Cycles() {
				fmt.Fprintln(cmd.ErrOrStderr(), "⚠ import cycle: "+strings.Join(c, " → "))
			}
			if diffFrom != "" {
				prefix, _ := s.Git.Run("rev-parse", "--show-prefix")
				before, err := NodesAt(s.Git, diffFrom, strings.TrimSpace(prefix), version, cmd.ErrOrStderr())
				if err != nil {
					return err
				}
				d := Diff(before, Nodes(s.Graph, nil))
				d.From, d.To = diffFrom, "the working tree"
				if format == "json" {
					enc := json.NewEncoder(cmd.OutOrStdout())
					enc.SetIndent("", "  ")
					return enc.Encode(d)
				}
				fmt.Fprint(cmd.OutOrStdout(), DiffText(d))
				return nil
			}
			g, marked := s.Graph, map[string]bool{}
			if len(focus) > 0 {
				names, err := s.Graph.Select(focus)
				if err != nil {
					return err
				}
				for _, n := range names {
					marked[n] = true
				}
				g = Around(s.Graph, names)
			}
			if onlyAffected {
				base := graphBase
				if base == "" {
					if base = affected.BaseFromEnv(); base == "" {
						base = s.Config.Release.Branch
					}
				}
				res := affected.Compute(s.Git, s.Graph, base)
				fmt.Fprintln(cmd.ErrOrStderr(), "affected: "+res.Reason)
				keep := map[string]bool{}
				for _, n := range res.Projects {
					if g.Get(n) != nil {
						keep[n] = true
					}
				}
				for _, f := range res.Changed {
					if owner := s.Graph.Owner(f); owner != nil {
						marked[owner.Name] = true
					}
				}
				g = Subgraph(g, keep)
			}
			var out string
			switch format {
			case "text":
				out = Text(g, marked)
			case "json":
				out = JSON(g, marked)
			case "dot":
				out = DOT(g, marked)
			case "html":
				out = HTML(g, s.Config.Workspace.Name, marked)
			case "tui":
				_, err := tea.NewProgram(NewBrowser(Nodes(g, marked)), tea.WithOutput(cmd.OutOrStdout())).Run()
				return err
			default:
				return fmt.Errorf("unknown format %q (text, json, dot, html, tui)", format)
			}
			if file != "" {
				if err := os.WriteFile(file, []byte(out), 0o644); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "wrote "+file)
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	graph.Flags().StringVar(&format, "format", "text", "text, json, dot, html or tui")
	graph.Flags().StringVarP(&file, "output", "o", "", "write to a file")
	graph.Flags().StringSliceVar(&focus, "focus", nil, "keep these projects and everything connected to them (names, globs, tag:x)")
	graph.Flags().BoolVar(&onlyAffected, "affected", false, "keep the projects affected by changes since --base; mark the changed ones")
	graph.Flags().StringVar(&graphBase, "base", "", "base branch for --affected (default: the PR target, else the release branch)")
	graph.Flags().StringVar(&diffFrom, "diff", "", "show how the graph changed since a revision (a branch, tag or commit): projects and imports added and removed, kinds and stacks changed; text or json")

	var projectsJSON bool
	projects := &cobra.Command{
		Use:   "projects",
		Short: "List the workspace's projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if projectsJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(Nodes(s.Graph, nil))
			}
			for _, n := range Nodes(s.Graph, nil) {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-8s %-7s %s\n", n.Name, n.Kind, n.Language, n.Root)
			}
			return nil
		},
	}
	projects.Flags().BoolVar(&projectsJSON, "json", false, "print JSON")

	var base string
	var affectedJSON, affectedPaths bool
	affectedCmd := &cobra.Command{
		Use:   "affected",
		Short: "List the projects affected by changes since a base branch",
		Long: `List the projects affected by changes since a base branch, one per line.

--json prints the whole answer, with "watch": for each app, the directories a
deployment platform should watch (the app's own and those of everything it
depends on, transitively), for Railway or Render watch paths and GitOps path
filters. --paths prints just that, one app per line: the app, a tab, then its
directories. Both list every app, affected or not.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if base == "" {
				if base = affected.BaseFromEnv(); base == "" {
					base = s.Config.Release.Branch
				}
			}
			res := affected.Compute(s.Git, s.Graph, base)
			watch := affected.WatchPaths(s.Graph)
			if affectedPaths {
				names := make([]string, 0, len(watch))
				for name := range watch {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", name, strings.Join(watch[name], " "))
				}
				return nil
			}
			if affectedJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(struct {
					affected.Result
					Watch map[string][]string `json:"watch,omitempty"`
				}{res, watch})
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", res.Reason)
			for _, p := range res.Projects {
				fmt.Fprintln(cmd.OutOrStdout(), p)
			}
			return nil
		},
	}
	affectedCmd.Flags().StringVar(&base, "base", "", "base branch (default: the PR target, else the release branch)")
	affectedCmd.Flags().BoolVar(&affectedJSON, "json", false, "print JSON, with the directories to watch for each app")
	affectedCmd.Flags().BoolVar(&affectedPaths, "paths", false, "print each app and the directories a deployment platform should watch")
	return []*cobra.Command{graph, projects, affectedCmd}
}
