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
	value, _ := AskOrBack(prompt, options, def, in, out, false)
	return value
}

// AskOrBack is Ask with one more answer: b (or back) says the person wants the
// question before this one. An option whose value or label is "b" or "back"
// keeps its meaning. The list shows the b line only when canBack is true; with
// canBack false it is Ask.
//
// Parameters: as Ask, and canBack is true when a question before this one exists.
// Returns: the Value of the answer, and true when the person asked to go back
// (the value is then "").
func AskOrBack(prompt string, options []Option, def string, in *bufio.Reader, out io.Writer, canBack bool) (string, bool) {
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
	if canBack {
		fmt.Fprintf(out, "  b) %-*s  %s\n", width, "Back", "the question before this one")
	}
	for range maxTries {
		fmt.Fprintf(out, "Choice [%s]: ", shown)
		line, err := in.ReadString('\n')
		typed := strings.TrimSpace(line)
		if value, ok := pick(options, def, typed); ok {
			return value, false
		}
		if canBack && (strings.EqualFold(typed, "b") || strings.EqualFold(typed, "back")) {
			return "", true
		}
		if err != nil {
			return def, false
		}
		fmt.Fprintf(out, "type a number from 1 to %d, or Enter for the default\n", len(options))
	}
	fmt.Fprintf(out, "Using the default, %s.\n", def)
	return def, false
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
