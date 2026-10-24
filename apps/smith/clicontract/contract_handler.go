package clicontract

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Commands are `smith commands`, `smith kinds` and `smith info`.
func Commands(version string) []*cobra.Command {
	var commandsJSON, kindsJSON, infoJSON bool
	commands := &cobra.Command{
		Use:   "commands",
		Short: "List smith's commands with their arguments and options (--json for editors)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			described := DescribeCommands(cmd.Root())
			if commandsJSON {
				return encode(cmd.OutOrStdout(), described)
			}
			for _, d := range described {
				fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-10s %s\n", d.Name, d.Group, d.Description)
			}
			return nil
		},
	}
	commands.Flags().BoolVar(&commandsJSON, "json", false, "print the commands as one JSON document")

	kinds := &cobra.Command{
		Use:   "kinds",
		Short: "List the project kinds smith add accepts (--json for editors)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			set, closeSet, err := plugins(version, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer closeSet()
			list := Kinds(set)
			if kindsJSON {
				return encode(cmd.OutOrStdout(), list)
			}
			for _, k := range list {
				fmt.Fprintf(cmd.OutOrStdout(), "%-16s %s\n", k.Kind, k.Description)
			}
			return nil
		},
	}
	kinds.Flags().BoolVar(&kindsJSON, "json", false, "print the kinds as one JSON document")

	info := &cobra.Command{
		Use:   "info",
		Short: "Show the installed smith, the newest released one, and this workspace's settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			report := WorkspaceReport(version, ".")
			if infoJSON {
				return encode(cmd.OutOrStdout(), report)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "smith %s\n", report.CLI.Version)
			if report.CLI.Latest != nil {
				fmt.Fprintf(out, "latest release: %s", *report.CLI.Latest)
				if report.CLI.UpdateAvailable {
					fmt.Fprint(out, " (an update is available)")
				}
				fmt.Fprintln(out)
			}
			if report.Workspace == nil {
				fmt.Fprintln(out, "not in a smith workspace")
				return nil
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: io.Discard})
			if err != nil {
				fmt.Fprintf(out, "workspace: %s (not loaded: %v)\n", report.Workspace.Root, err)
				return nil
			}
			defer s.Close()
			fmt.Fprintln(out)
			fmt.Fprint(out, WorkspaceSummary(report.Workspace.Root, s.Config, s.Graph))
			return nil
		},
	}
	info.Flags().BoolVar(&infoJSON, "json", false, "print the report as one JSON document")
	schema := &cobra.Command{
		Use:   "schema",
		Short: "Print the JSON Schema of smith.toml, for editors",
		Long: `Print the JSON Schema of smith.toml. Editors that read TOML schemas (Even Better
TOML in VS Code) pick it up from the #:schema line smith new writes at the top
of smith.toml; it is served at ` + workspace.SchemaURL + `.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := cmd.OutOrStdout().Write(workspace.Schema)
			return err
		},
	}
	return []*cobra.Command{commands, kinds, info, schema}
}

// plugins are the workspace's plugins, or the built-in ones outside a
// workspace.
func plugins(version string, log io.Writer) (*pluginhost.PluginSet, func(), error) {
	if _, err := workspace.FindRoot("."); err == nil {
		s, err := session.Open(session.Options{Dir: ".", Version: version, Log: log})
		if err != nil {
			return nil, nil, err
		}
		return s.Plugins, s.Close, nil
	}
	set, err := pluginhost.Start(pluginhost.Options{Config: workspace.Config{}, Version: version, Log: log})
	if err != nil {
		return nil, nil, err
	}
	return set, set.Close, nil
}

func encode(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
