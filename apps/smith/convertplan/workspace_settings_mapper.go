package convertplan

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// WorkspaceSettings derives smith.toml's [workspace], [ci] and [release]
// sections. hasFile reports whether a workspace-relative file exists;
// modules maps Go project roots to their module paths.
func WorkspaceSettings(w *nxworkspace.Workspace, dirName string, hasFile func(string) bool, modules map[string]string) workspace.Config {
	cfg := workspace.Config{Schema: workspace.SchemaVersion}
	name, scope := splitPackageName(w.RootPackage.Name)
	if name == "" || name == "source" {
		name = dirName
	}
	cfg.Workspace.Name, cfg.Workspace.Scope = name, scope
	if legacy := w.Nx.Mnci; legacy != nil {
		if legacy.WorkspaceName != "" {
			cfg.Workspace.Name = legacy.WorkspaceName
		}
		if legacy.Scope != "" {
			cfg.Workspace.Scope = legacy.Scope
		}
		for _, ci := range legacy.CI {
			switch ci {
			case "github", "azure":
				if !contains(cfg.CI.Providers, ci) {
					cfg.CI.Providers = append(cfg.CI.Providers, ci)
				}
			case "both":
				cfg.CI.Providers = []string{"github", "azure"}
			}
		}
	}
	if len(cfg.CI.Providers) == 0 {
		if hasFile(".github/workflows/ci.yml") || hasFile(".github/workflows/ci.yaml") {
			cfg.CI.Providers = append(cfg.CI.Providers, "github")
		}
		if hasFile("azure-pipelines.yml") {
			cfg.CI.Providers = append(cfg.CI.Providers, "azure")
		}
	}
	cfg.Workspace.ModulePrefix = modulePrefix(w, modules)
	if w.Nx.DefaultBase != "" && w.Nx.DefaultBase != "main" {
		cfg.Release.Branch = w.Nx.DefaultBase
	}
	cfg.Release.GitHubReleases = w.Nx.Release.CreatesGitHubReleases()
	return cfg
}

// splitPackageName turns "@acme/source" into ("source", "@acme").
func splitPackageName(name string) (base, scope string) {
	if strings.HasPrefix(name, "@") {
		scope, base, _ = strings.Cut(name, "/")
		return base, scope
	}
	return name, ""
}

// modulePrefix is the Go module prefix: @nx-go's modulePrefix option, else
// what the module paths have before their project directories
// (github.com/acme/repo for github.com/acme/repo/apps/api).
func modulePrefix(w *nxworkspace.Workspace, modules map[string]string) string {
	for _, p := range w.Nx.Plugins {
		if p.Plugin != "@nx-go/nx-go" {
			continue
		}
		var prefix string
		if raw, ok := p.Options["modulePrefix"]; ok && json.Unmarshal(raw, &prefix) == nil && prefix != "" {
			return prefix
		}
	}
	for root, module := range modules {
		if prefix, ok := strings.CutSuffix(module, "/"+root); ok {
			return prefix
		}
	}
	return ""
}
