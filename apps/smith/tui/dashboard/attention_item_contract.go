package dashboard

// Kind says what an item of the attention queue is about.
type Kind string

// The kinds of item, in the order the queue lists them.
const (
	KindEmpty    Kind = "empty"
	KindChecking Kind = "checking"
	KindDoctor   Kind = "doctor"
	KindSetup    Kind = "setup"
	KindRun      Kind = "run"
	KindWarnings Kind = "warnings"
	KindChanges  Kind = "changes"
	KindRelease  Kind = "release"
	KindUpdates  Kind = "updates"
	// KindCompletion is the offer to set up tab completion; it only waits.
	KindCompletion Kind = "completion"
	// KindRemote is the state of the remote cache and the entries it lacks.
	KindRemote Kind = "remote"
	KindGood       Kind = "good"
)

// Severity is how much an item asks for attention; the page draws its icon in
// the colour of it.
type Severity int

// The severities, from the calmest to the loudest.
const (
	// SeverityMuted is something that waits and costs nothing.
	SeverityMuted Severity = iota
	// SeveritySuccess is good news.
	SeveritySuccess
	// SeverityInfo is something worth a look.
	SeverityInfo
	// SeverityWarning is work that the changes left.
	SeverityWarning
	// SeverityFailure is something broken.
	SeverityFailure
)

// Item is one line of the attention queue: what needs a look and what to press.
type Item struct {
	Kind     Kind
	Severity Severity
	// Title is the sentence; Detail is an optional short second line.
	Title, Detail string
	// Buttons are the things to do, the first one first.
	Buttons []Button
}

// Button is a thing to do: the label, the id of the command a press runs and
// the key that does the same ("" when there is none on Home).
type Button struct {
	Label     string
	CommandID string
	Key       string
}
