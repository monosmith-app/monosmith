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
}

// Dialogue is how the setup talks to the person. The terminal of `smith cache
// setup` and the screens of the interactive shell each implement it, so both
// show the same decisions from the same engine. Every method may block until
// the person answers.
type Dialogue interface {
	// Say tells the person something; it asks nothing.
	Say(text string)
	// Choose asks for one option; def is the Value of the default. It returns that Value.
	Choose(prompt string, options []Option, def string) string
	// Text asks for a typed line and returns it trimmed.
	Text(prompt string) string
	// Confirm asks a yes or no question; it returns true only for yes.
	Confirm(c Confirmation) bool
	// Working says that slow work starts (text says what). The returned
	// function says it ended.
	Working(text string) (done func())
	// Show presents a titled page of lines and returns when the person moves on.
	Show(title string, lines []string)
	// Secret copies a secret unseen and offers the choices after it (see
	// cachekey.Session). produce is nil for a key made by smith.
	Secret(noun string, produce func() (string, error), actions []cachekey.KeyAction) error
}
