package pluginhost

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginenv"
	"github.com/monosmith-app/monosmith/apps/smith/plugininstall"
	"github.com/monosmith-app/monosmith/apps/smith/pluginpack"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
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

// yesNo prints a flag as a table cell.
func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// ExternalStatuses describes the external plugins smith.toml declares; a plugin
// smith cannot even resolve is a status with its problem.
func ExternalStatuses(root string, cfg workspace.Config) []pluginadmit.Status {
	store, err := plugintrust.OpenStore()
	if err != nil {
		return nil
	}
	var out []pluginadmit.Status
	names := make([]string, 0, len(cfg.Plugins.External))
	for name := range cfg.Plugins.External {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p, err := ResolveExternal(root, cfg.Plugins.External, name)
		if err != nil {
			out = append(out, pluginadmit.Status{Plugin: pluginadmit.Plugin{Name: name, Source: workspace.FileName}, Problem: err.Error()})
			continue
		}
		out = append(out, pluginadmit.Describe(store, root, []pluginadmit.Plugin{p})...)
	}
	return out
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
		Short: "List the built-in plugins, the external ones the workspace uses (trusted or not) and the ones installed on the plugin search path",
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
				// The rest is set for an external plugin: Used is true when
				// the workspace declares it, and Env, Trusted and Pinned say
				// what it may see and whether smith may start it.
				Used    bool     `json:"used,omitempty"`
				Command []string `json:"command,omitempty"`
				Env     []string `json:"env,omitempty"`
				Trusted *bool    `json:"trusted,omitempty"`
				Pinned  *bool    `json:"pinned,omitempty"`
				Problem string   `json:"problem,omitempty"`
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
			var statuses []pluginadmit.Status
			if root != "" {
				if cfg, err := workspace.Load(root); err == nil {
					statuses = ExternalStatuses(root, cfg)
				}
			}
			used := map[string]bool{}
			for _, st := range statuses {
				used[st.Name] = true
				trusted, pinned := st.Trusted, st.Pinned
				rows = append(rows, row{Name: st.Name, Source: st.Source, Used: true, Command: st.Command, Env: st.Env, Trusted: &trusted, Pinned: &pinned, Problem: st.Problem})
			}
			for _, p := range installed {
				if !used[p.Name] {
					rows = append(rows, row{Name: p.Name, Version: p.Version, Source: p.Dir, Env: pluginenv.Visible(p.Env, nil)})
				}
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
			fmt.Fprintf(out, "\nExternal plugins this workspace uses (smith starts them only when trusted):\n")
			if len(statuses) == 0 {
				fmt.Fprintln(out, "  none")
			} else {
				fmt.Fprintf(out, "  %-14s %-18s %-8s %-7s %s\n", "PLUGIN", "FROM", "TRUSTED", "PINNED", "MAY SEE BEYOND THE BASELINE")
			}
			for _, st := range statuses {
				fmt.Fprintf(out, "  %-14s %-18s %-8s %-7s %s\n", st.Name, st.Source, yesNo(st.Trusted), yesNo(st.Pinned), orDash(strings.Join(st.Env, ", ")))
				if st.Problem != "" {
					fmt.Fprintf(out, "      %s\n", st.Problem)
				}
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
	cmd.AddCommand(list, checkCommand(version), packCommand(), installCommand())
	cmd.AddCommand(trustCommands()...)
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

// packCommand is `smith plugin pack`, reading manifests the way the host does.
func packCommand() *cobra.Command {
	return pluginpack.Command(func(dir string) (pluginpack.Plugin, error) {
		m, err := readInstalled(dir)
		if err != nil {
			return pluginpack.Plugin{}, err
		}
		return pluginpack.Plugin{Name: m.Name, Version: m.Version, Command: m.Command}, nil
	})
}

// installCommand is `smith plugin install`, reading manifests the way the host does.
func installCommand() *cobra.Command {
	return plugininstall.Command(func(dir string) (plugininstall.Described, error) {
		m, err := readInstalled(dir)
		if err != nil {
			return plugininstall.Described{}, err
		}
		return plugininstall.Described{Plugin: plugininstall.Plugin{Name: m.Name, Version: m.Version}, Argv: m.Argv, Command0: m.Command[0]}, nil
	})
}
