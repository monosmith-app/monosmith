package choiceprompt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// maxTries is how many times one question is asked before its default is taken.
const maxTries = 3

// Option is one answer to a question: a Label, the Value Ask returns, and a
// Hint that says in plain words what the answer does.
type Option struct{ Label, Value, Hint string }

// Ask prints a question with numbered answers on out and reads the choice from
// in. Enter takes def. A closed input takes def as well, so a pipe never blocks.
// A number, a value or a label (any case) picks an answer.
//
// Parameters: prompt is the question; options are the answers; def is the
// Value Enter takes; in and out are the terminal.
// Returns: the Value of the answer.
func Ask(prompt string, options []Option, def string, in *bufio.Reader, out io.Writer) string {
	fmt.Fprintf(out, "\n%s\n", prompt)
	width := 0
	for _, o := range options {
		width = max(width, len(o.Label))
	}
	shown := def
	for i, o := range options {
		line := fmt.Sprintf("  %d) %-*s", i+1, width, o.Label)
		if o.Hint != "" {
			line += "  " + o.Hint
		}
		fmt.Fprintln(out, strings.TrimRight(line, " "))
		if o.Value == def {
			shown = strconv.Itoa(i + 1)
		}
	}
	for range maxTries {
		fmt.Fprintf(out, "Choice [%s]: ", shown)
		line, err := in.ReadString('\n')
		if value, ok := pick(options, def, strings.TrimSpace(line)); ok {
			return value
		}
		if err != nil {
			return def
		}
		fmt.Fprintf(out, "type a number from 1 to %d, or Enter for the default\n", len(options))
	}
	fmt.Fprintf(out, "Using the default, %s.\n", def)
	return def
}

// pick turns what a person typed into the value of an answer; ok is false for
// text that matches none.
func pick(options []Option, def, typed string) (string, bool) {
	if typed == "" {
		return def, true
	}
	if n, err := strconv.Atoi(typed); err == nil && n >= 1 && n <= len(options) {
		return options[n-1].Value, true
	}
	for _, o := range options {
		if strings.EqualFold(typed, o.Value) || strings.EqualFold(typed, o.Label) {
			return o.Value, true
		}
	}
	return "", false
}
