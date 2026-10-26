package settingsview

// Backend is what the tab calls to read and change things. Every function
// does what the matching `smith` command does, in this process; the model
// runs them off the update loop and shows what they return. A nil function
// means the thing is not offered (Install, until `smith plugin install`
// exists in this smith).
type Backend struct {
	// Load reads the workspace.
	Load func() Snapshot
	// SetPlugin switches a built-in plugin on or off (smith plugin enable|disable).
	SetPlugin func(name string, enable bool) error
	// SetArchitecture chooses how the workspace's code is laid out (smith architecture).
	SetArchitecture func(name string) error
	// Review measures an external plugin for a trust decision (the review of
	// smith plugin trust).
	Review func(name string) (TrustReview, error)
	// Trust records the decision, and with pin writes the sha256 to
	// smith.toml; Untrust forgets it. They return what to tell the person.
	Trust   func(name string, pin bool) (string, error)
	Untrust func(name string) (string, error)
	// SetSetting and UnsetSetting are smith config set and unset; they check
	// the values against the schema and return what to tell the person.
	SetSetting   func(path string, values []string) (string, error)
	UnsetSetting func(path string) (string, error)
	// Install installs a plugin by name or path (smith plugin install). It
	// stays nil until that command exists; the Plugins section then shows the
	// command as a hint, and turns into a live action when it is set.
	Install func(source string) (string, error)
	// Style is the person's saved choice of theme and symbols.
	Style StylePrefs
	// TipSeen and MarkTipSeen keep whether Home still points at this tab.
	TipSeen     bool
	MarkTipSeen func()
}

// TrustReview is what the person reads before trusting a plugin.
type TrustReview struct {
	// Text is the review as `smith plugin trust` prints it.
	Text string
	// Trusted is true when the person's decision on file still matches.
	Trusted bool
}

// StylePrefs is the interface preferences the Style section edits.
type StylePrefs struct {
	// Theme and Icons are what is saved: "" or auto, or a name (see package theme).
	Theme, Icons string
	// Path is the file the choice is kept in, for the section to name.
	Path string
	// Save keeps the choice in the smith home.
	Save func(theme, icons string) error
	// Getenv reads the environment variables that win over the saved choice.
	Getenv func(string) string
	// Dark is whether the terminal's background is dark.
	Dark bool
}
