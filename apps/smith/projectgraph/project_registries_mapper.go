package projectgraph

import (
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ProjectRegistries are the registries a project installs from and publishes
// to, from smith.toml: installs from [workspace] registry; publishing to
// [projects.<name>] registry, else [release] registry, else the install
// registry. Nil when smith.toml names none, so plugins keep the public
// registries.
func ProjectRegistries(cfg workspace.Config, project string) *protocol.Registries {
	install := registry(cfg, cfg.Workspace.Registry)
	publishName := cfg.Projects[project].Registry
	if publishName == "" {
		publishName = cfg.Release.Registry
	}
	publish := install
	if publishName != "" {
		publish = registry(cfg, publishName)
	}
	none := publishName == workspace.NoRegistry
	if install == nil && publish == nil && !none {
		return nil
	}
	return &protocol.Registries{Install: install, Publish: publish, NoPublish: none}
}

func registry(cfg workspace.Config, name string) *protocol.Registry {
	r, ok := cfg.Registries[name]
	if name == "" || !ok {
		return nil
	}
	return &protocol.Registry{Name: name, Auth: r.Auth, TokenEnv: r.TokenEnv, NPM: r.NPM, PyPI: r.PyPI, PyPIUpload: r.PyPIUpload, NuGet: r.NuGet}
}
