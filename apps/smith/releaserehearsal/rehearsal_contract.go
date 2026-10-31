package releaserehearsal

// Ecosystems a rehearsal knows how to ask.
const (
	NPM    = "npm"
	PyPI   = "pypi"
	NuGet  = "nuget"
	Crates = "crates.io"
)

// Result states of one check.
const (
	Pass = "pass"
	Warn = "warn"
	Fail = "fail"
)

// Check names, one per question a rehearsal asks of a registry.
const (
	CheckRegistry   = "registry"
	CheckCredential = "credential"
	CheckAccess     = "access"
	CheckIdentity   = "identity"
)

// Target is one registry a release publishes one project to.
type Target struct {
	Project   string `json:"project"`
	Ecosystem string `json:"ecosystem"`
	// URL is the registry endpoint the publish uses.
	URL string `json:"url"`
	// Auth is the sign-in kind smith.toml gives the registry (azure-artifacts,
	// github-packages or token); empty for a public registry or a feed named
	// only in .npmrc.
	Auth string `json:"auth,omitempty"`
	// TokenEnv is the variable smith.toml names for Auth "token".
	TokenEnv string `json:"tokenEnv,omitempty"`
	// Names are the packages the release publishes there.
	Names []string `json:"names"`
	// NpmrcToken is the credential an .npmrc line gives this registry, with
	// its ${VAR} expanded; it is never printed or encoded.
	NpmrcToken string `json:"-"`
	// NpmrcVar is the variable that .npmrc line names, for the messages.
	NpmrcVar string `json:"-"`
}

// Result is the answer to one check of one target.
type Result struct {
	Project string `json:"project"`
	// Registry is the host of the target's URL.
	Registry string `json:"registry"`
	Check    string `json:"check"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	// Fix is the exact command or setting that resolves a failure or warning.
	Fix string `json:"fix,omitempty"`
}

// Report is everything a rehearsal found, in the order it asked.
type Report struct {
	Results []Result `json:"results"`
	Passed  int      `json:"passed"`
	Failed  int      `json:"failed"`
	Warned  int      `json:"warnings"`
}

// Add records a result and counts it.
func (r *Report) Add(res Result) {
	r.Results = append(r.Results, res)
	switch res.Status {
	case Pass:
		r.Passed++
	case Fail:
		r.Failed++
	default:
		r.Warned++
	}
}
