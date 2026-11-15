package plugintrust

import "time"

// Executable is a file a plugin's command runs, with its digest.
type Executable struct {
	// Path is the absolute path the command resolved to, symlinks followed.
	Path string `json:"path"`
	// SHA256 is the hex digest of the file's content.
	SHA256 string `json:"sha256"`
}

// Fingerprint is what a plugin's command line is made of: the command as
// given, and the files it runs. The first executable is the program; for an
// interpreter (`node plugin.mjs`) the script follows it.
type Fingerprint struct {
	Command     []string     `json:"command"`
	Executables []Executable `json:"executables"`
}

// Record is one trust decision, kept in the user's smith home.
type Record struct {
	// Workspace is the workspace root the decision was made for.
	Workspace string `json:"workspace"`
	// Plugin is the plugin's name in that workspace.
	Plugin string `json:"plugin"`
	Fingerprint
	TrustedAt time.Time `json:"trustedAt"`
}

// Verdict is whether a plugin may start, and if not, why.
type Verdict struct {
	Trusted bool `json:"trusted"`
	// Via is how a trusted plugin got there: "recorded" (the user trusted
	// it) or "pinned" (in CI, smith.toml pins its digest).
	Via string `json:"via,omitempty"`
	// Reason is one of the Reason constants for a plugin that is not trusted.
	Reason Reason `json:"reason,omitempty"`
	// Detail is a short human sentence about the reason.
	Detail string `json:"detail,omitempty"`
}

// Reason says why a plugin is not trusted.
type Reason string

// The reasons a plugin is not trusted.
const (
	ReasonNew            Reason = "new"
	ReasonCommandChanged Reason = "command-changed"
	ReasonFileChanged    Reason = "file-changed"
	ReasonNotPinned      Reason = "not-pinned"
	ReasonPinMismatch    Reason = "pin-mismatch"
)
