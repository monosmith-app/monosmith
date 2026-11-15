// Package releaselineage decides whether a project's release tags can be
// trusted before a release: tags left under an old name, a version published
// with no tag, and tags behind what the registry already has. Each would make
// the next release start from the wrong version.
package releaselineage
