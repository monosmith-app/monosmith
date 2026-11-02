package internallink

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Option is one answer to a question, with what it does in plain words.
type Option struct {
	Label   string
	Meaning string
}

// Asker puts a real choice to the user. The first option is the default. A nil
// Asker means smith has no terminal to ask in.
type Asker interface {
	// Choose shows the question and the options and returns the index of the
	// option picked. It fails when the answer cannot be read.
	Choose(question string, options []Option) (int, error)
}

// TerminalAsker asks on a terminal: the question, each option with its
// meaning, and a prompt that takes a number or Enter for the default.
type TerminalAsker struct {
	in  *bufio.Reader
	out io.Writer
}

// NewTerminalAsker reads answers from in and writes questions to out.
func NewTerminalAsker(in io.Reader, out io.Writer) *TerminalAsker {
	return &TerminalAsker{in: bufio.NewReader(in), out: out}
}

// Choose implements Asker. A wrong answer is asked again.
func (t *TerminalAsker) Choose(question string, options []Option) (int, error) {
	fmt.Fprintln(t.out, question)
	for i, o := range options {
		mark := ""
		if i == 0 {
			mark = " (default)"
		}
		fmt.Fprintf(t.out, "  %d) %s%s\n", i+1, o.Label, mark)
		if o.Meaning != "" {
			fmt.Fprintf(t.out, "     %s\n", o.Meaning)
		}
	}
	for {
		fmt.Fprintf(t.out, "Choose 1-%d [1]: ", len(options))
		line, err := t.in.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" && err != nil {
			return 0, fmt.Errorf("no answer was given: %w", err)
		}
		if line == "" {
			return 0, nil
		}
		if n, convErr := strconv.Atoi(line); convErr == nil && n >= 1 && n <= len(options) {
			return n - 1, nil
		}
		fmt.Fprintf(t.out, "Type a number from 1 to %d, or press Enter for the default.\n", len(options))
	}
}
