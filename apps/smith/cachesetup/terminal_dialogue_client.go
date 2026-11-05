package cachesetup

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
)

// Terminal is the Dialogue of a terminal: numbered choices and typed lines.
type Terminal struct {
	// In and Out are the terminal.
	In  *bufio.Reader
	Out io.Writer
	// SecretFn runs the clipboard session of a secret.
	SecretFn func(noun string, produce func() (string, error), actions []cachekey.KeyAction) error
}

// Say prints a line.
func (t Terminal) Say(text string) { fmt.Fprintln(t.Out, text) }

// Choose asks with numbered answers.
func (t Terminal) Choose(prompt string, options []Option, def string) string {
	return choiceprompt.Ask(prompt, options, def, t.In, t.Out)
}

// Text asks for a line.
func (t Terminal) Text(prompt string) string {
	fmt.Fprintf(t.Out, "\n%s: ", prompt)
	line, _ := t.In.ReadString('\n')
	return strings.TrimSpace(line)
}

// Confirm prints what will happen, the exact command or change, then asks.
func (t Terminal) Confirm(c Confirmation) bool {
	if c.Text != "" {
		fmt.Fprintf(t.Out, "\n%s\n", c.Text)
	}
	if c.Shown != "" {
		for _, l := range strings.Split(c.Shown, "\n") {
			fmt.Fprintf(t.Out, "  %s\n", l)
		}
	}
	return choiceprompt.Ask(c.Prompt, []Option{{Label: c.Yes, Value: "yes", Hint: c.YesHint}, {Label: c.No, Value: "no", Hint: c.NoHint}}, "yes", t.In, t.Out) == "yes"
}

// Working prints nothing: the command was shown before it ran.
func (t Terminal) Working(string) func() { return func() {} }

// Show prints the page.
func (t Terminal) Show(title string, lines []string) {
	fmt.Fprintf(t.Out, "\n== %s ==\n", title)
	for _, l := range lines {
		fmt.Fprintln(t.Out, l)
	}
}

// Secret runs the clipboard session.
func (t Terminal) Secret(noun string, produce func() (string, error), actions []cachekey.KeyAction) error {
	return t.SecretFn(noun, produce, actions)
}

var _ Dialogue = Terminal{}
