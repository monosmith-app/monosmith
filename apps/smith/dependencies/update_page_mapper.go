package dependencies

import (
	"github.com/monosmith-app/monosmith/apps/smith/registrypage"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// pageOf is the address of an update's package on the registry its first
// declaring project installs from ("" when it cannot be known, see package
// registrypage). A declaration in the root manifest takes the registry of the
// first project the plugin serves that names one.
//
// Parameters: u is the update; projects are the projects of its plugin;
// ambientPrivate is true when the workspace points the tool at a private
// registry outside smith.toml.
func pageOf(u Update, projects []protocol.Project, ambientPrivate bool) string {
	var install *protocol.Registry
	pick := func(p protocol.Project) {
		if install == nil && p.Registries != nil {
			install = p.Registries.Install
		}
	}
	for _, s := range u.Sites {
		for _, p := range projects {
			if s.Project == "" || p.Name == s.Project {
				pick(p)
			}
		}
	}
	return registrypage.Page(u.Plugin, u.Name, install, ambientPrivate)
}
