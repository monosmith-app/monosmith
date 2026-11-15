package settingsview

// Backend is what the tab calls to read and change things. Every function
// does what the matching `smith` command does, in this process; the model
// runs them off the update loop and shows what they return. What needs the
// network or a download (smith plugin install) is not here: the tab asks the
// shell to run that command line, which shows it first (Result.Run).
type Backend struct {
	// Load reads the workspace.
	Load func() Snapshot
	// SetPlugin switches a built-in plugin on or off (smith plugin enable|disable).
	// It returns a warning to show with the change ("" for none).
	SetPlugin func(name string, enable bool) (string, error)
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
	// Style is the person's saved choice of theme and symbols.
	Style StylePrefs
	// Update is what the Update section shows and changes.
	Update UpdatePrefs
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

// UpdateInfo is what smith knows about a newer smith.
type UpdateInfo struct {
	// Current is the running version; Latest the newest on the channel ("" when
	// the registry has not been asked); Command is what installs it.
	Current, Latest, Command string
	// Available is true when Latest is newer; Runnable when smith can run Command
	// (it knows how it was installed).
	Available, Runnable bool
	// Muted is true when the person asked not to be reminded in the interface.
	Muted bool
}

// UpdatePrefs is the Update section's access to smith update.
type UpdatePrefs struct {
	// Info is what smith knows now, without a request.
	Info UpdateInfo
	// Check asks the registry now and returns what it found.
	Check func() (UpdateInfo, error)
	// SetMuted keeps whether the interface leaves the person alone about updates.
	SetMuted func(muted bool) error
}
