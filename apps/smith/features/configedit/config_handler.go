package configedit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/monosmith-app/monosmith/apps/smith/features/operation"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/apps/smith/shared/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/shared/marks"
	"github.com/spf13/cobra"
)

// ProjectLoader reports what a project's plugin says about the project
// called name, in the workspace at dir; log takes what the plugins print.
// It returns an error when the workspace has no such project.
type ProjectLoader func(dir, name string, log io.Writer) (ProjectFacts, error)

// Command is `smith config`.
//
// Parameters: project reports a project's facts for -p, --project; nil means
// this smith cannot name projects, and -p is refused.
// Returns: the command.
func Command(project ProjectLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and change settings of smith.toml",
		Long: `Read and change the settings of smith.toml's [workspace], [ci], [release],
[cache], [plugins], [agent] and [dependencies] tables. Values are checked against smith.toml's schema,
and only the setting's value changes in the file: comments and every other
line stay as they are. A default stack is set per language, as in
default_stacks.node. With -p, --project the commands read and change the
settings of one project ([projects.<name>]); a value its manifest owns is
shown read-only. Edit stacks, conventions and tasks in the file itself.`,
		Example: `  smith config list
  smith config set default_stacks.node node-jest
  smith config set ci.devcontainer false
  smith config set ci.providers github azure
  smith config unset release.branch
  smith config list -p web
  smith config set -p web tag_prefix web-v`,
	}
	root := func() (string, error) { return workspace.FindRoot(".") }
	var projectName string
	// path is the setting an argument names: a project's when -p is given.
	path := func(cmd *cobra.Command, dir, arg string) (string, error) {
		if projectName == "" && !strings.HasPrefix(arg, projectsPrefix) {
			return arg, nil
		}
		full := arg
		if projectName != "" {
			full = ProjectPath(projectName, arg)
		}
		if taskOutputsPath.MatchString(full) {
			return full, nil // a task's outputs, which Unset reaches without a field
		}
		if name, _, ok := splitProjectPath(strings.TrimPrefix(full, projectsPrefix)); ok && project != nil {
			// Writing [projects.<name>] for a project that does not exist would be silent.
			if _, err := project(dir, name, cmd.ErrOrStderr()); err != nil {
				return "", err
			}
		}
		return full, nil
	}

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
			if projectName != "" {
				return listProject(cmd, dir, projectName, project, asJSON)
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
				text := r.Description
				if r.Reason != "" {
					text += " Why: " + r.Reason
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Path, value, text)
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
			setting, err := path(cmd, dir, args[0])
			if err != nil {
				return err
			}
			value, set, err := Get(dir, setting)
			if err != nil {
				return err
			}
			if !set {
				return fmt.Errorf("%s is not set; it has its default (see https://github.com/monosmith-app/monosmith/blob/main/docs/reference/configuration.md)", setting)
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		},
	}

	set := &cobra.Command{
		Use:   "set <setting> <value...>",
		Short: "Change one setting (a list takes several values, or one comma-separated)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := root()
			if err != nil {
				return err
			}
			setting, err := path(cmd, dir, args[0])
			if err != nil {
				return err
			}
			if err := Set(dir, setting, args[1:]); err != nil {
				return err
			}
			value, _, _ := Get(dir, setting)
			fmt.Fprintf(cmd.OutOrStdout(), marks.Passed+" %s = %s\n", setting, value)
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
			setting, err := path(cmd, dir, args[0])
			if err != nil {
				return err
			}
			if err := Unset(dir, setting); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), marks.Passed+" %s removed: it has its default\n", setting)
			return nil
		},
	}
	operation.Wrap(set, operation.Spec{Describe: func(_ *cobra.Command, args []string) string { return "set " + args[0] }})
	operation.Wrap(unset, operation.Spec{Describe: func(_ *cobra.Command, args []string) string { return "unset " + args[0] }})
	for _, c := range []*cobra.Command{list, get, set, unset} {
		commandflags.AddOneProjectFlag(c, &projectName, "the project whose settings ([projects.<name>]) to read or change")
	}
	cmd.AddCommand(list, get, set, unset)
	return cmd
}

// listProject prints the settings of one project: the values its manifest
// owns, read-only with the file, then the settings of [projects.<name>].
func listProject(cmd *cobra.Command, dir, name string, project ProjectLoader, asJSON bool) error {
	if project == nil {
		return errors.New("this smith cannot read the projects of a workspace\n→ edit [projects." + TableKey(name) + "] in smith.toml")
	}
	facts, err := project(dir, name, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	view, err := ReadProjectSettings(dir, facts)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(view)
	}
	fmt.Fprintf(out, "%s: values from %s (change them there)\n", view.Name, view.Manifest)
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for _, f := range view.Facts {
		fmt.Fprintf(w, "  %s\t%s\n", f.Key, f.Value)
	}
	_ = w.Flush()
	fmt.Fprintf(out, "\nSettings of [projects.%s] in %s\n", TableKey(view.Name), workspace.FileName)
	for _, st := range view.Settings {
		value := st.Value
		if !st.Set {
			value = "(default)"
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\n", st.Key(), value, st.Explanation())
	}
	return w.Flush()
}
