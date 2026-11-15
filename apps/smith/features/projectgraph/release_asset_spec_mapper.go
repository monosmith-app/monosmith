package projectgraph

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// AssetSpec is what a plugin needs to build an app's release assets, or nil
// for a project without them.
func (p *Project) AssetSpec() *protocol.AssetSpec {
	r := p.Release
	if r == nil || r.Build != "" {
		// A build command makes its own files; no plugin builds archives.
		return nil
	}
	return &protocol.AssetSpec{
		Name:            r.AssetName,
		Product:         r.Product,
		Platforms:       r.Platforms,
		OSAlias:         r.OSAlias,
		ArchAlias:       r.ArchAlias,
		VersionVariable: r.VersionVariable,
		OutDir:          "dist/release/" + p.Name,
	}
}
