package release

import (
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PublishesToNoRegistry reports whether smith.toml says the project goes to no
// registry: registry = "none" (the tag is the release, JitPack and Packagist
// read it) or registry = "github-release" (the package file is attached to the
// GitHub Release). Such a project asks for no registry credential, is never
// looked up on a registry, and is never dispatched to pub.dev.
func PublishesToNoRegistry(p *projectgraph.Project) bool {
	return p.Registries != nil && (p.Registries.NoPublish || p.Registries.PackToRelease)
}

// PacksToRelease reports whether the release packs the project and attaches
// the package file to its GitHub Release: a package with registry =
// "github-release". An app is never packed, and a native package is published
// by its native jobs.
func PacksToRelease(p *projectgraph.Project) bool {
	return p.Kind == protocol.ProjectPackage && !Native(p) && p.Registries != nil && p.Registries.PackToRelease
}

// PackDir is the workspace-relative folder a pack writes a project's package
// files to; it sits beside the project's release journal.
func PackDir(project string) string {
	return "dist/release/" + project + "/packages"
}
