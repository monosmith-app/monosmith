package convert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ConfigText renders smith.toml for a converted workspace: only the settings
// that differ from smith's defaults, in a stable order.
func ConfigText(cfg workspace.Config) string {
	var b strings.Builder
	b.WriteString("#:schema " + workspace.SchemaURL + "\n")
	b.WriteString("# smith workspace configuration. Written by `smith convert`; edit it freely or with the TUI.\n")
	b.WriteString("# Every setting is optional except [workspace] name; see https://github.com/monosmith-app/monosmith/blob/main/docs/reference/configuration.md.\n")
	fmt.Fprintf(&b, "schema = %d\n\n[workspace]\nname = %s\n", workspace.SchemaVersion, q(cfg.Workspace.Name))
	if cfg.Workspace.Scope != "" {
		fmt.Fprintf(&b, "scope = %s\n", q(cfg.Workspace.Scope))
	}
	if cfg.Workspace.ModulePrefix != "" {
		fmt.Fprintf(&b, "module_prefix = %s\n", q(cfg.Workspace.ModulePrefix))
	}
	if len(cfg.Workspace.Exclude) > 0 {
		b.WriteString("# Not projects: standalone folders outside the npm workspaces.\n")
		fmt.Fprintf(&b, "exclude = %s\n", list(cfg.Workspace.Exclude))
	}
	if len(cfg.CI.Providers) > 0 || cfg.CI.AzurePool != "" || len(cfg.CI.AzureVariableGroups) > 0 {
		b.WriteString("\n[ci]\n")
		if len(cfg.CI.Providers) > 0 {
			fmt.Fprintf(&b, "providers = %s\n", list(cfg.CI.Providers))
		}
		if cfg.CI.AzurePool != "" {
			fmt.Fprintf(&b, "azure_pool = %s\n", q(cfg.CI.AzurePool))
		}
		if len(cfg.CI.AzureVariableGroups) > 0 {
			fmt.Fprintf(&b, "azure_variable_groups = %s\n", list(cfg.CI.AzureVariableGroups))
		}
	}
	if cfg.Release.Branch != "" || cfg.Release.Remote != "" || cfg.Release.GitHubReleases || cfg.Release.Registry != "" {
		b.WriteString("\n[release]\n")
		if cfg.Release.Branch != "" {
			fmt.Fprintf(&b, "branch = %s\n", q(cfg.Release.Branch))
		}
		if cfg.Release.Remote != "" {
			fmt.Fprintf(&b, "remote = %s\n", q(cfg.Release.Remote))
		}
		if cfg.Release.GitHubReleases {
			b.WriteString("github_releases = true\n")
		}
		if cfg.Release.Registry != "" {
			fmt.Fprintf(&b, "registry = %s\n", q(cfg.Release.Registry))
		}
	}
	for _, name := range sortedKeys(cfg.Registries) {
		r := cfg.Registries[name]
		fmt.Fprintf(&b, "\n[registries.%s]\nauth = %s\n", key(name), q(r.Auth))
		if r.NPM != "" {
			fmt.Fprintf(&b, "npm = %s\n", q(r.NPM))
		}
	}
	if len(cfg.DefaultStacks) > 0 {
		b.WriteString("\n[default_stacks]\n")
		for _, lang := range sortedKeys(cfg.DefaultStacks) {
			fmt.Fprintf(&b, "%s = %s\n", key(lang), q(cfg.DefaultStacks[lang]))
		}
	}
	for _, name := range sortedKeys(cfg.Conventions) {
		c := cfg.Conventions[name]
		fmt.Fprintf(&b, "\n[conventions.%s]\n", key(name))
		if len(c.Languages) > 0 {
			fmt.Fprintf(&b, "languages = %s\n", list(c.Languages))
		}
		if len(c.Exclude) > 0 {
			fmt.Fprintf(&b, "exclude = %s\n", list(c.Exclude))
		}
	}
	writeTasks(&b, "tasks", cfg.Tasks)
	for _, g := range cfg.TaskGroups {
		b.WriteString(workspace.TaskGroupText(g))
	}
	for _, name := range sortedKeys(cfg.Projects) {
		p := cfg.Projects[name]
		if p.TagPrefix != "" {
			fmt.Fprintf(&b, "\n[projects.%s]\ntag_prefix = %s\n", key(name), q(p.TagPrefix))
		}
		if r := p.Release; r != nil {
			fmt.Fprintf(&b, "\n[projects.%s.release]\n", key(name))
			if r.AssetName != "" {
				fmt.Fprintf(&b, "asset_name = %s\n", q(r.AssetName))
			}
			if r.Product != "" {
				fmt.Fprintf(&b, "product = %s\n", q(r.Product))
			}
			if len(r.Platforms) > 0 {
				fmt.Fprintf(&b, "platforms = %s\n", list(r.Platforms))
			}
			if len(r.OSAlias) > 0 {
				fmt.Fprintf(&b, "os_alias = %s\n", table(r.OSAlias))
			}
			if len(r.ArchAlias) > 0 {
				fmt.Fprintf(&b, "arch_alias = %s\n", table(r.ArchAlias))
			}
			if r.VersionVariable != "" {
				fmt.Fprintf(&b, "version_variable = %s\n", q(r.VersionVariable))
			}
		}
		writeTasks(&b, "projects."+key(name)+".tasks", p.Tasks)
	}
	return b.String()
}

func writeTasks(b *strings.Builder, prefix string, tasks map[string]workspace.TaskOverride) {
	for _, name := range sortedKeys(tasks) {
		fmt.Fprintf(b, "\n[%s.%s]\n", prefix, key(name))
		for _, line := range workspace.TaskFieldLines(tasks[name]) {
			b.WriteString(line + "\n")
		}
	}
}

// q quotes a TOML basic string.
func q(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`, "\r", `\r`)
	return `"` + r.Replace(s) + `"`
}

// key quotes a table key when it is not a bare key.
func key(s string) string {
	for _, c := range s {
		bare := c == '-' || c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if !bare {
			return q(s)
		}
	}
	return s
}

func list(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = q(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func table(m map[string]string) string {
	parts := make([]string, 0, len(m))
	for _, k := range sortedKeys(m) {
		parts = append(parts, key(k)+" = "+q(m[k]))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
