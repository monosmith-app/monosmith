package protocol

// Publishes reports whether a release publishes the project to a registry: a
// package, unless smith.toml says registry = "none". Plugins list the
// credentials a release needs only for such projects.
func (p Project) Publishes() bool {
	return p.Kind == ProjectPackage && (p.Registries == nil || !p.Registries.NoPublish)
}
