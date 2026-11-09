package release

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// NativeTag marks a project built on each OS (smith ci native), whose release
// assets each native job attaches for its own platform.
const NativeTag = "native"

// Native reports whether a project is built on each OS.
func Native(p *projectgraph.Project) bool { return slices.Contains(p.Tags, NativeTag) }

// attachesItsFile reports whether a project's release has an asset without
// the project having a [release] table: a recipe attaches its own file, so a
// URL can pin a version of it. A resumed release lists it from the plugin.
func attachesItsFile(p *projectgraph.Project) bool { return p.Language == "recipe" }

// AssetSpec is what a plugin needs to build an app's release assets, or nil
// for a project without them.
func AssetSpec(p *projectgraph.Project) *protocol.AssetSpec {
	return p.AssetSpec()
}
