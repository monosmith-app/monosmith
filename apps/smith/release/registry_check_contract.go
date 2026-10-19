package release

// RegistryTarget is one registry a release publishes a project to, with
// what the registry must accept.
type RegistryTarget struct {
	Project string
	// Registry is npm, PyPI, NuGet or RubyGems.
	Registry string
	// Names are the packages the release publishes there.
	Names   []string
	Version string
}

// RegistryFinding is what a registry check found. A blocking finding stops
// the release before anything is published; the others are said and the
// release goes on.
type RegistryFinding struct {
	Project  string `json:"project"`
	Registry string `json:"registry"`
	Message  string `json:"message"`
	// Fix says what to do about it.
	Fix      string `json:"fix,omitempty"`
	Blocking bool   `json:"blocking"`
}

// String is the finding on one line, for the refusal and the warnings.
func (f RegistryFinding) String() string {
	text := f.Registry + " (" + f.Project + "): " + f.Message
	if f.Fix != "" {
		text += ": " + f.Fix
	}
	return text
}
