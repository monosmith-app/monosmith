package configedit

import (
	"encoding/json"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Command is `smith config`.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and change settings of smith.toml",
		Long: `Read and change the settings of smith.toml's [workspace], [ci], [release],
[cache] and [plugins] tables. Values are checked against smith.toml's schema,
and only the setting's value changes in the file: comments and every other
line stay as they are. A default stack is set per language, as in
default_stacks.node. Edit stacks, conventions, tasks and projects in the file
itself.`,
		Example: `  smith config list
  smith config set default_stacks.node node-jest
  smith config set ci.devcontainer false
  smith config set ci.providers github azure
  smith config unset release.branch`,
	}
	root := func() (string, error) { return workspace.FindRoot(".") }

	var asJSON bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List the settings, their values and what they do",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := root()
			if err != nil {
				return err
			}
			fields, err := Fields()
			if err != nil {
				return err
			}
			type row struct {
				Field
				Value string `json:"value,omitempty"`
				Set   bool   `json:"set"`
			}
			cfg, err := workspace.Load(dir)
			if err != nil {
				return err
			}
			// A default stack exists per language; list the ones the file sets.
			var languages []string
			for language := range cfg.DefaultStacks {
				languages = append(languages, language)
			}
			sort.Strings(languages)
			for _, language := range languages {
				fields = append(fields, DefaultStackField(language))
			}
			var rows []row
			for _, f := range fields {
				value, set, err := Get(dir, f.Path)
				if err != nil {
					return err
				}
				rows = append(rows, row{Field: f, Value: value, Set: set})
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			for _, r := range rows {
				value := r.Value
				if !r.Set {
					value = "(default)"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Path, value, r.Description)
			}
			return w.Flush()
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "print the settings as JSON")

	get := &cobra.Command{
		Use:   "get <setting>",
		Short: "Print one setting's value as written in smith.toml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := root()
			if err != nil {
				return err
			}
			value, set, err := Get(dir, args[0])
			if err != nil {
				return err
			}
			if !set {
				return fmt.Errorf("%s is not set; it has its default (see https://github.com/monosmith-app/monosmith/blob/main/docs/reference/configuration.md)", args[0])
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}

	set := &cobra.Command{
		Use:   "set <setting> <value...>",
		Short: "Change one setting (a list takes several values)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := root()
			if err != nil {
				return err
			}
			if err := Set(dir, args[0], args[1:]); err != nil {
				return err
			}
			value, _, _ := Get(dir, args[0])
			fmt.Fprintf(cmd.OutOrStdout(), "✔ %s = %s\n", args[0], value)
			return nil
		},
	}

	unset := &cobra.Command{
		Use:   "unset <setting>",
		Short: "Remove one setting, so it takes its default",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := root()
			if err != nil {
				return err
			}
			if err := Unset(dir, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✔ %s removed: it has its default\n", args[0])
			return nil
		},
	}
	cmd.AddCommand(list, get, set, unset)
	return cmd
}
