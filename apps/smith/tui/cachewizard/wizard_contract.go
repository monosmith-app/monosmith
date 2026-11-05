package cachewizard

import "github.com/monosmith-app/monosmith/apps/smith/cachesetup"

// Kind is what the screen of the wizard asks or shows.
type Kind int

// The screens.
const (
	// Idle: smith works between two questions; the screen shows a spinner.
	Idle Kind = iota
	// Working: a program runs (a cloud listing, a sign-in), with its words.
	Working
	// Choose: one option of a list.
	Choose
	// Text: a typed line.
	Text
	// Confirm: a yes or no that shows the exact command or change first.
	Confirm
	// Page: lines to read, with Enter to go on.
	Page
	// Secret: the key or token was copied unseen; buttons act on it.
	Secret
)

// Button is one choice of the secret screen, a key press with a label.
type Button struct {
	Key   rune
	Label string
}

// Screen is what the engine asks the person, as data the model draws.
type Screen struct {
	Kind Kind
	// Prompt is the question, or the title of a page.
	Prompt string
	// Options and Default are the choices of a Choose screen.
	Options []cachesetup.Option
	Default string
	// Confirmation is the question of a Confirm screen.
	Confirmation cachesetup.Confirmation
	// Lines are the lines of a Page.
	Lines []string
	// Noun names the secret ("Key", "SAS token") and Buttons are its choices.
	Noun    string
	Buttons []Button
}

// Event is news from the engine, which runs off the interface's goroutine.
type Event interface{ event() }

type (
	// SayEvent is a line the engine says.
	SayEvent struct{ Text string }
	// AskEvent replaces the screen.
	AskEvent struct{ Screen Screen }
	// IdleEvent says the slow work ended; the engine goes on.
	IdleEvent struct{}
	// EndEvent says the engine ended: Remote is what the person chose, Err why
	// it stopped, Cancelled that the person closed it.
	EndEvent struct {
		Remote    string
		Err       error
		Cancelled bool
	}
)

func (SayEvent) event()  {}
func (AskEvent) event()  {}
func (IdleEvent) event() {}
func (EndEvent) event()  {}

// Reply is the answer to a screen.
type Reply struct {
	// Value is the chosen option or the typed line.
	Value string
	// Yes is the answer of a Confirm screen.
	Yes bool
}

// Outcome is what the host must do after a key or a click.
type Outcome struct {
	// Reply, when set, answers the screen the engine waits on.
	Reply *Reply
	// Press is a key for the secret screen's session (0 when none).
	Press rune
	// Cancel closes the wizard.
	Cancel bool
}
