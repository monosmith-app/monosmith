package pluginhost

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
	"github.com/spf13/cobra"
)

// orDash prints an empty column as "-".
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// Command is `smith plugin`: list plugins, and serve a built-in one over stdio.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{Use: "plugin", Short: "Inspect plugins, or serve a built-in plugin"}
	cmd.AddCommand(&cobra.Command{
		Use:    "serve <name>",
		Short:  "Serve a built-in plugin on stdin/stdout (smith runs this itself)",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(_ *cobra.Command, args []string) error {
			ctor, ok := Builtin(args[0])
			if !ok {
				return fmt.Errorf("no built-in plugin %q", args[0])
			}
			return serve.Serve(ctor(version), os.Stdin, os.Stdout)
		},
	})
	var asJSON bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List the built-in plugins and the ones installed on the plugin search path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			type row struct {
				Name         string               `json:"name"`
				Version      string               `json:"version,omitempty"`
				Source       string               `json:"source"`
				Roles        []protocol.Role      `json:"roles,omitempty"`
				Provides     []string             `json:"provides,omitempty"`
				Requires     []protocol.PluginRef `json:"requires,omitempty"`
				Conflicts    []protocol.PluginRef `json:"conflicts,omitempty"`
				Capabilities any                  `json:"capabilities,omitempty"`
			}
			var rows []row
			for _, name := range append(BuiltinNames(), sortedNames(BuiltinAddons)...) {
				ctor, _ := Builtin(name)
				p := ctor(version)
				info, err := p.Describe(protocol.InitializeParams{})
				if err != nil {
					return err
				}
				rows = append(rows, row{Name: name, Source: "built-in", Roles: info.EffectiveRoles(), Provides: info.Provides, Requires: info.Requires, Conflicts: info.Conflicts, Capabilities: serve.CapabilitiesOf(p)})
			}
			root, _ := workspace.FindRoot(".")
			installed, err := FindInstalled(root)
			if err != nil {
				return err
			}
			for _, p := range installed {
				rows = append(rows, row{Name: p.Name, Version: p.Version, Source: p.Dir})
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			fmt.Fprintf(out, "%-16s %-28s %-20s %s\n", "PLUGIN", "LEVELS", "SLOTS", "REQUIRES")
			for _, r := range rows {
				if r.Source != "built-in" {
					continue
				}
				roles := make([]string, len(r.Roles))
				for i, role := range r.Roles {
					roles[i] = string(role)
				}
				needs := make([]string, 0, len(r.Requires)+len(r.Conflicts))
				for _, ref := range r.Requires {
					needs = append(needs, ref.String())
				}
				for _, ref := range r.Conflicts {
					needs = append(needs, "not "+ref.String())
				}
				fmt.Fprintf(out, "%-16s %-28s %-20s %s\n", r.Name, strings.Join(roles, ","), orDash(strings.Join(r.Provides, ",")), orDash(strings.Join(needs, "; ")))
			}
			fmt.Fprintf(out, "\nInstalled (search path: %s):\n", strings.Join(SearchPath(root), ", "))
			if len(installed) == 0 {
				fmt.Fprintln(out, "  none")
			}
			for _, p := range installed {
				fmt.Fprintf(out, "  %-14s %-10s %s\n", p.Name, p.Version, p.Dir)
			}
			return nil
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	cmd.AddCommand(list, checkCommand(version))
	for _, enable := range []bool{true, false} {
		verb := map[bool]string{true: "enable", false: "disable"}[enable]
		cmd.AddCommand(&cobra.Command{
			Use:   verb + " <name>...",
			Short: strings.ToUpper(verb[:1]) + verb[1:] + " built-in plugins in smith.toml",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(c *cobra.Command, args []string) error {
				root, err := workspace.FindRoot(".")
				if err != nil {
					return err
				}
				cfg, err := workspace.Load(root)
				if err != nil {
					return err
				}
				enabled, err := Toggle(cfg.Plugins.Enabled, args, enable)
				if err != nil {
					return err
				}
				if err := workspace.SetEnabledPlugins(root, enabled); err != nil {
					return err
				}
				fmt.Fprintf(c.OutOrStdout(), "enabled plugins: %s\n", strings.Join(enabled, ", "))
				return nil
			},
		})
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "set <name>...",
		Short: "Enable exactly these built-in plugins in smith.toml",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			enabled, err := Toggle([]string{args[0]}, args, true)
			if err != nil {
				return err
			}
			if err := workspace.SetEnabledPlugins(root, enabled); err != nil {
				return err
			}
			fmt.Fprintf(c.OutOrStdout(), "enabled plugins: %s\n", strings.Join(enabled, ", "))
			return nil
		},
	})
	return cmd
}

// Toggle returns the enabled list after enabling or disabling names. An empty
// current list means every built-in plugin.
func Toggle(current, names []string, enable bool) ([]string, error) {
	if len(current) == 0 {
		current = BuiltinNames()
	}
	set := map[string]bool{}
	for _, n := range current {
		set[n] = true
	}
	for _, n := range names {
		if _, ok := Builtins[n]; !ok {
			return nil, fmt.Errorf("no built-in plugin %q (built-in: %s)", n, strings.Join(BuiltinNames(), ", "))
		}
		set[n] = enable
	}
	var out []string
	for _, n := range BuiltinNames() {
		if set[n] {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one plugin must stay enabled")
	}
	return out, nil
}
