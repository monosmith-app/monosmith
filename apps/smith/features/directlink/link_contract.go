package directlink

import "time"

// Kind is what a smith entry on PATH is, which decides how it is pointed at the program.
type Kind string

// The kinds of entry.
const (
	// CmdShim is a Windows batch file (smith.cmd).
	CmdShim Kind = "cmd"
	// Ps1Shim is a PowerShell script (smith.ps1).
	Ps1Shim Kind = "ps1"
	// ShShim is a shell script (smith) that starts a launcher.
	ShShim Kind = "sh"
	// Symlink is a symbolic link to a launcher script.
	Symlink Kind = "symlink"
	// Launcher is a small program (a pip console script) that starts a language runtime.
	Launcher Kind = "launcher"
	// Direct is an entry that already starts the program.
	Direct Kind = "direct"
)

// Entry is one file named smith on PATH that belongs to a smith installation.
type Entry struct {
	// Path is the file.
	Path string `json:"path"`
	// Kind is what it is.
	Kind Kind `json:"kind"`
	// Backup is where the original was kept before it was changed: a file, or for a
	// symlink its old target. Empty for an entry that was not changed.
	Backup string `json:"backup,omitempty"`
}

// State is what smith remembers in ~/.smith/link.json.
type State struct {
	// Version is the smith version that was linked.
	Version string `json:"version"`
	// Installed is the program the package manager put in place.
	Installed string `json:"installed"`
	// Target is the program the entries call now.
	Target string `json:"target"`
	// Entries are the entries that were changed, with how to restore them.
	Entries []Entry `json:"entries"`
	// At is when the link was made.
	At time.Time `json:"at"`
}

// Options say where smith is and what it may touch.
type Options struct {
	// Home is the smith home (~/.smith); "" does nothing.
	Home string
	// Exe is the program that runs now: the one the package manager installed, or a
	// link of it in the run folder when smith already starts directly.
	Exe string
	// Version is the running smith.
	Version string
	// GOOS is runtime.GOOS.
	GOOS string
	// Now is the clock; time.Now when nil.
	Now func() time.Time
}

// Report says what Link did.
type Report struct {
	// Target is the program the entries call now.
	Target string
	// Changed are the entries that now call the program.
	Changed []Entry
	// Already are the entries that did already.
	Already []Entry
	// Skipped says why an entry was left alone ("path: reason").
	Skipped []string
}

// installed is the program the package manager put in place.
func (o Options) installed() string { return InstalledExe(o.Home, o.Exe) }
