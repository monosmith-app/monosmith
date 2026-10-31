package stackview

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// row is one stack as smith stacks reports it.
type row struct {
	pluginhost.CatalogStack
	// Default lists the languages whose default stack this is.
	Default []string `json:"default,omitempty"`
	// Projects use the stack.
	Projects []string `json:"projects,omitempty"`
}

// Command is `smith stacks`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "stacks",
		Short: "List the stacks projects can be built with, and who uses each",
		Long: `A stack is a named set of plugins a project is built with: a language plugin,
and the toolchain, structure, scaffold and deploy plugins on top. Plugins
publish stacks; smith.toml declares more under [stacks.<name>].

  smith add app web --stack node-vitest`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			var rows []row
			for _, st := range s.Stacks.All() {
				r := row{CatalogStack: st}
				for lang, name := range s.Config.DefaultStacks {
					if name == st.Name {
						r.Default = append(r.Default, lang)
					}
				}
				for _, p := range s.Graph.Projects {
					if p.Stack == st.Name {
						r.Projects = append(r.Projects, p.Name)
					}
				}
				rows = append(rows, r)
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			if len(rows) == 0 {
				fmt.Fprintln(out, "No stacks: no plugin publishes one and smith.toml declares none.")
				return nil
			}
			for _, r := range rows {
				fmt.Fprintf(out, "%s  (%s)", r.Name, r.Source)
				if len(r.Default) > 0 {
					fmt.Fprintf(out, "  default for %s", strings.Join(r.Default, ", "))
				}
				fmt.Fprintln(out)
				if r.Description != "" {
					fmt.Fprintf(out, "  %s\n", r.Description)
				}
				fmt.Fprintf(out, "  language   %s\n", r.Language)
				for _, part := range []struct {
					role  string
					names []string
				}{{"toolchain", r.Toolchain}, {"structure", r.Structure}, {"scaffold", nonEmpty(r.Scaffold)}, {"deploy", r.Deploy}} {
					if len(part.names) > 0 {
						fmt.Fprintf(out, "  %-10s %s\n", part.role, strings.Join(part.names, ", "))
					}
				}
				if len(r.Choices) > 0 {
					fmt.Fprintf(out, "  %-10s %s\n", "choices", strings.Join(r.Choices, ", "))
				}
				if r.Problem != "" {
					fmt.Fprintf(out, "  ✖ cannot be used: %s\n", r.Problem)
				}
				if len(r.Projects) > 0 {
					fmt.Fprintf(out, "  projects   %s\n", strings.Join(r.Projects, ", "))
				}
				fmt.Fprintln(out)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}

func nonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}
