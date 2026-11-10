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
// Used through a pointer, it is also BackAware: a choice then offers b for the
// question before. Used as a value, it never offers b.
type Terminal struct {
	// In and Out are the terminal.
	In  *bufio.Reader
	Out io.Writer
	// SecretFn runs the clipboard session of a secret.
	SecretFn func(noun, variable string, produce func() (string, error), actions []cachekey.KeyAction) error
	// canBack is set by AllowBack for the next question.
	canBack bool
}

// AllowBack implements BackAware.
func (t *Terminal) AllowBack(can bool) { t.canBack = can }

// ask is a numbered choice; b panics with BackRequest when a way back was allowed.
func (t Terminal) ask(prompt string, options []Option, def string) string {
	value, back := choiceprompt.AskOrBack(prompt, options, def, t.In, t.Out, t.canBack)
	if back {
		panic(BackRequest{})
	}
	return value
}

// Say prints a line.
func (t Terminal) Say(text string) { fmt.Fprintln(t.Out, text) }

// Choose asks with numbered answers.
func (t Terminal) Choose(prompt string, options []Option, def string) string {
	return t.ask(prompt, options, def)
}

// Text asks for a line; an empty line takes def.
func (t Terminal) Text(prompt, def string) string {
	if def != "" {
		fmt.Fprintf(t.Out, "\n%s [%s]: ", prompt, def)
	} else {
		fmt.Fprintf(t.Out, "\n%s: ", prompt)
	}
	line, _ := t.In.ReadString('\n')
	if line = strings.TrimSpace(line); line == "" {
		return def
	}
	return line
}

// Confirm prints what will happen, the exact command or change, then asks.
func (t Terminal) Confirm(c Confirmation) (yes, commit bool) {
	if c.Text != "" {
		fmt.Fprintf(t.Out, "\n%s\n", c.Text)
	}
	if c.Shown != "" {
		for _, l := range strings.Split(c.Shown, "\n") {
			fmt.Fprintf(t.Out, "  %s\n", l)
		}
	}
	def := "yes"
	if c.Earlier != nil && !c.Earlier.Yes {
		def = "no"
	}
	return t.ask(c.Prompt, []Option{{Label: c.Yes, Value: "yes", Hint: c.YesHint}, {Label: c.No, Value: "no", Hint: c.NoHint}}, def) == "yes", false
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
func (t Terminal) Secret(noun, variable string, produce func() (string, error), actions []cachekey.KeyAction) error {
	return t.SecretFn(noun, variable, produce, actions)
}

var (
	_ Dialogue  = Terminal{}
	_ BackAware = (*Terminal)(nil)
)
