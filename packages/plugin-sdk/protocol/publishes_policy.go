package protocol

// Publishes reports whether a release publishes the project to a registry: a
// package, unless smith.toml says registry = "none" or the manifest says it is
// never published (PublishTo "none", a pubspec's publish_to: none). Plugins list the
// credentials a release needs only for such projects.
func (p Project) Publishes() bool {
	return p.Kind == ProjectPackage && p.PublishTo != "none" && (p.Registries == nil || !p.Registries.NoPublish)
}
