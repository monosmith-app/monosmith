package npmpropagation

import "time"

// Default values of the wait.
const (
	// DefaultMax is the longest wait when nothing is configured.
	DefaultMax = 15 * time.Minute
	// DefaultConfirmations is how many consecutive full probes must pass.
	DefaultConfirmations = 3
	// DefaultSpacing is the pause between the confirming probes.
	DefaultSpacing = 30 * time.Second
	// DefaultFirstDelay is the first pause of the backoff; it doubles up to DefaultMaxDelay.
	DefaultFirstDelay = 2 * time.Second
	// DefaultMaxDelay is the longest pause of the backoff.
	DefaultMaxDelay = 30 * time.Second
	// DefaultProgressEvery is how often a progress line is printed.
	DefaultProgressEvery = time.Minute
)

// Registry asks a package registry the way a person's npm does.
type Registry interface {
	// Lookup reports whether name@version exists, and the URL of its tarball.
	Lookup(name, version string) (found bool, tarball string)
	// Fetchable reports whether a HEAD request of the URL succeeds.
	Fetchable(url string) bool
}

// Clock tells the time and pauses; tests replace it so they never sleep.
type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

// Config tunes Await. A zero field takes its default, except Max, where zero
// means do not wait: one probe only.
type Config struct {
	// Max is the longest wait in all.
	Max time.Duration
	// Confirmations is how many consecutive full probes must pass, Spacing apart.
	Confirmations int
	Spacing       time.Duration
	// FirstDelay and MaxDelay bound the exponential backoff while packages are missing.
	FirstDelay, MaxDelay time.Duration
	// ProgressEvery is how often Progress is called while waiting.
	ProgressEvery time.Duration
	// Progress receives one line; nothing is printed when nil.
	Progress func(line string)
}

// Outcome is what Await found.
type Outcome struct {
	// Ready is true when every package was fetchable on every confirming probe.
	Ready bool
	// Missing are the packages the last probe could not fetch (empty when
	// the wait ended before the confirming probes were all done).
	Missing []string
	// Waited is how long Await took.
	Waited time.Duration
}
