package cachesetup

import (
	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
)

// Option is one answer of a choice.
type Option = choiceprompt.Option

// Confirmation is a question that runs a command or writes a file when the
// answer is yes. Shown is the exact command or change, so what is confirmed is
// what happens.
type Confirmation struct {
	// Text says what will happen, in plain words; may be empty.
	Text string
	// Shown is the exact command line or file change; may be empty.
	Shown string
	// Prompt is the question.
	Prompt string
	// Yes and No are the two answers, each with a hint that says what it does.
	Yes, YesHint, No, NoHint string
	// OfferCommit asks an interface to show the unticked box "Commit this change
	// so it can be undone" with the question; the answer says whether it was
	// ticked. The terminal has the --commit flag instead and never ticks it.
	OfferCommit bool
	// Earlier is the answer the person gave before going back to this
	// question; an interface marks it again. nil when the question is new.
	Earlier *ConfirmAnswer
}

// ConfirmAnswer is the answer to a Confirmation.
type ConfirmAnswer struct {
	// Yes is true for the yes answer; Commit is true when the commit box was ticked.
	Yes, Commit bool
}

// Dialogue is how the setup talks to the person. The terminal of `smith cache
// setup` and the screens of the interactive shell each implement it, so both
// show the same decisions from the same engine. Every method may block until
// the person answers. Choose, Text, Confirm and Show may panic with
// BackRequest when the person asks for the question before (see BackAware).
type Dialogue interface {
	// Say tells the person something; it asks nothing.
	Say(text string)
	// Choose asks for one option; def is the Value of the default. It returns that Value.
	Choose(prompt string, options []Option, def string) string
	// Text asks for a typed line and returns it trimmed; def is the line to
	// start from ("" for none), kept when the person only confirms.
	Text(prompt, def string) string
	// Confirm asks a yes or no question; yes is true only for yes, and commit
	// says that the box offered by OfferCommit was ticked.
	Confirm(c Confirmation) (yes, commit bool)
	// Working says that slow work starts (text says what). The returned
	// function says it ended.
	Working(text string) (done func())
	// Show presents a titled page of lines and returns when the person moves on.
	Show(title string, lines []string)
	// Secret copies a secret unseen and offers the choices after it (see
	// cachekey.Session). produce is nil for a key made by smith.
	Secret(noun, variable string, produce func() (string, error), actions []cachekey.KeyAction) error
}
