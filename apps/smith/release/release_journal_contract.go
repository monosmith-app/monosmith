package release

import "time"

// Journal is the record of what a project's releases published and where:
// dist/release/<project>/journal.json. It only grows. A registry never lets a
// published version be taken back, so it is not for undoing a release: it
// shows what is done, which the resume path and `smith release --status` use.
type Journal struct {
	Project string         `json:"project"`
	Entries []JournalEntry `json:"entries"`
}

// JournalEntry is one step of a release, once it is done or has failed.
type JournalEntry struct {
	Time    time.Time `json:"time"`
	Version string    `json:"version"`
	// Step is publish (the plugin's registry), package (an npm, PyPI, NuGet
	// or RubyGems package), image, tag, push, mirror, github-release,
	// pub-dispatch or release (the step that failed).
	Step string `json:"step"`
	// Target is where it went: the registry, package, image, tag or URL.
	Target string `json:"target"`
	// Outcome is published, existing (the registry already had it), done or
	// failed.
	Outcome string `json:"outcome"`
	Error   string `json:"error,omitempty"`
}

// Journal outcomes.
const (
	OutcomePublished = "published"
	OutcomeExisting  = "existing"
	OutcomeDone      = "done"
	OutcomeFailed    = "failed"
)
