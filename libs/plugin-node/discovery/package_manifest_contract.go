package discovery

// PackageManifest is the subset of package.json the plugin reads.
type PackageManifest struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Private              bool              `json:"private"`
	Scripts              map[string]string `json:"scripts"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	// PeerDependenciesMeta marks peers optional; npm and Yarn also take a
	// name listed only here as an optional peer of any version.
	PeerDependenciesMeta map[string]struct {
		Optional bool `json:"optional"`
	} `json:"peerDependenciesMeta"`
	Engines struct {
		VSCode string `json:"vscode"`
	} `json:"engines"`
	Smith struct {
		Kind  string `json:"kind"`
		Stack string `json:"stack"`
		// Native: the project is built on each OS (smith ci native), because
		// what it ships differs per platform (prebuilt addons, a bundled
		// runtime, a per-platform VS Code extension).
		Native bool `json:"native"`
	} `json:"smith"`
}
