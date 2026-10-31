package protocol

// Method names of protocol version 1. A host never calls a method the plugin
// did not advertise in its Capabilities.
const (
	MethodInitialize = "initialize"
	MethodShutdown   = "shutdown"
	MethodDiscover   = "discover"
	MethodGenerate   = "generate"
	MethodVersion    = "version/write"
	MethodPublish    = "publish"
	MethodCI         = "ci"
	MethodDoctor     = "doctor"
	// MethodCancel is a notification: the host asks the plugin to stop the
	// request with the given id. It is sent only to a plugin that advertised
	// the cancel capability.
	MethodCancel = "cancel"
	// MethodContribute asks a convention plugin for its additions.
	MethodContribute = "contribute"
	// MethodRemove lets a plugin clean up after a project is removed.
	MethodRemove = "remove"
	// MethodChangeKind asks what changing a project's kind edits.
	MethodChangeKind = "project/kind"
	// MethodPublished asks for the newest version a registry has of a
	// project, so release can check its tags against it.
	MethodPublished = "published"
	// MethodInstall adds packages to a project, or installs a workspace.
	MethodInstall = "dependencies/install"
	// MethodUninstall removes packages from a project.
	MethodUninstall = "dependencies/uninstall"
	// MethodDeclared lists the dependencies manifests declare.
	MethodDeclared = "dependencies/declared"
	// MethodRewrite changes declared specs in place.
	MethodRewrite = "dependencies/rewrite"
	// MethodLatest asks registries for the newest releases (smith up).
	MethodLatest = "dependencies/latest"
	// MethodComponents lists the components a project ships, from its
	// lockfile or manifest, for the release's SBOM.
	MethodComponents = "sbom/components"
	// MethodDerive asks for the files a plugin derives from facts and the
	// project graph (smith derive).
	MethodDerive = "derive"
)
