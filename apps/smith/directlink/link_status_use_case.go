package directlink

// Status says how smith starts now.
type Status struct {
	// Version is the running smith.
	Version string
	// Direct is true when this process runs from a link: no launcher is in front of it.
	Direct bool
	// Installed is the program the package manager installed.
	Installed string
	// Entries are the smith entries on PATH, each with its kind; Direct means it
	// starts the program without a launcher.
	Entries []Entry
	// Linked is the version that was linked last; "" when none was.
	Linked string
}

// Check reads how smith starts now, and changes nothing.
//
// Parameters: o says where smith is.
// Returns: the status.
func Check(o Options) Status {
	s := Status{Version: o.Version, Direct: RunningFromLink(o.Home, o.Exe), Installed: InstalledExe(o.Home, o.Exe)}
	if state, ok := LoadState(o.Home); ok {
		s.Linked = state.Version
	}
	s.Entries = FindEntries(s.Installed, o.GOOS, RunDir(o.Home), expectedTarget(o))
	return s
}

// Needed is true when the first run of this version should link: smith runs
// behind a launcher, no link was made for this version, and an entry exists
// that would change. It reads two small files and never writes.
//
// Parameters: o says where smith is.
// Returns: true when Link would change something.
func Needed(o Options) bool {
	if o.Home == "" || RunningFromLink(o.Home, o.Exe) {
		return false
	}
	if state, ok := LoadState(o.Home); ok && state.Version == o.Version {
		return false
	}
	for _, e := range FindEntries(o.installed(), o.GOOS, RunDir(o.Home), expectedTarget(o)) {
		if e.Kind != Direct {
			return true
		}
	}
	return false
}
