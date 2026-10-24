package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// argQuestion asks for one positional argument a person left out.
type argQuestion struct {
	// Label names what is asked: "Kind", "Name of the project".
	Label string
	// Options, when set, are listed and may be answered by number.
	Options []string
	// Validate rejects an answer; its message is shown and the question asked again.
	Validate func(answer string) error
}

// maxArgTries is how many times one question is asked before giving up.
const maxArgTries = 3

// isInteractive reports whether a person is at a terminal that can answer: not
// a pipe, and not CI.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

// askMissingArguments completes args with the answers to the questions for
// the arguments not given (questions[i] is for args[i]). An empty line or
// end of input is an error, never a guess.
func askMissingArguments(args []string, questions []argQuestion, in io.Reader, out io.Writer) ([]string, error) {
	completed := append([]string(nil), args...)
	reader := bufio.NewReader(in)
	for i := len(args); i < len(questions); i++ {
		q := questions[i]
		answer, err := askArgument(reader, out, q)
		if err != nil {
			return nil, err
		}
		completed = append(completed, answer)
	}
	return completed, nil
}

func askArgument(reader *bufio.Reader, out io.Writer, q argQuestion) (string, error) {
	for i, o := range q.Options {
		fmt.Fprintf(out, "  %d) %s\n", i+1, o)
	}
	var last error
	for attempt := 0; attempt < maxArgTries; attempt++ {
		fmt.Fprintf(out, "%s: ", q.Label)
		line, readErr := reader.ReadString('\n')
		answer := strings.TrimSpace(line)
		if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(q.Options) {
			answer = q.Options[n-1]
		}
		switch {
		case answer == "":
			last = fmt.Errorf("%s is needed", strings.ToLower(q.Label))
		case q.Validate != nil && q.Validate(answer) != nil:
			last = q.Validate(answer)
		default:
			return answer, nil
		}
		if readErr != nil {
			break
		}
		fmt.Fprintf(out, "%v\n", last)
	}
	return "", last
}

// newQuestions are asked by `smith new` for what it was not given.
func newQuestions() []argQuestion {
	return []argQuestion{{Label: "Name of the workspace", Validate: func(a string) error { return ValidateName("workspace", a) }}}
}

// addQuestions are asked by `smith add` for what it was not given. The
// language is asked later, by chooseLanguage, as it is when only --lang is
// missing.
func addQuestions() []argQuestion {
	return []argQuestion{
		{Label: "Kind (app is deployed, lib is internal, package is published)", Options: []string{"app", "lib", "package"}, Validate: func(a string) error {
			if _, _, combined := SplitKind(a); combined || a == "app" || a == "lib" || a == "package" {
				return nil
			}
			return fmt.Errorf("%q is not a kind: app, lib or package (smith kinds lists the combined ones)", a)
		}},
		{Label: "Name of the project", Validate: func(a string) error { return ValidateName("project", a) }},
	}
}

// argsOrAsk accepts the arguments of a command like commandargs.Exactly, but
// where a person can answer (interactive says so) fewer are fine: the rest
// are asked for.
func argsOrAsk(spec commandargs.Spec, interactive func() bool) cobra.PositionalArgs {
	exact := commandargs.Exactly(spec)
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < len(spec.Names) && interactive() {
			return nil
		}
		return exact(cmd, args)
	}
}
