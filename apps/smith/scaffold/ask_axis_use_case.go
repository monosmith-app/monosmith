package scaffold

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// maxAxisRounds bounds the ask loop: each round answers one axis.
const maxAxisRounds = 8

// CompleteWith asks for the axes that the choices of smith add --with need and
// that have no default (the trigger and the runtime of a function app, ADR 0072).
// It never picks for the person: with no terminal it returns the choices as they
// are and the error that names the options, so the add stops before any write.
//
// Parameters: s is the open session; opts the options of the add; in and out the
// terminal; interactive says a person is there to answer.
// Returns: the choices with the answers added, and the solver's error when a
// person cannot answer or the answer ends in a different problem.
func CompleteWith(s *session.Session, opts AddOptions, in io.Reader, out io.Writer, interactive bool) ([]string, error) {
	with := append([]string(nil), opts.With...)
	if len(with) == 0 {
		return with, nil
	}
	language := ""
	if opts.Lang != "" {
		asked, err := s.Plugins.ForLanguage(opts.Lang)
		if err != nil {
			return with, nil
		}
		language = asked.Info.Plugin.Name
	}
	reader := bufio.NewReader(in)
	for range maxAxisRounds {
		_, err := choiceselect.SolveIn(s.Plugins.Offers(), language, with)
		var need *choiceselect.NeedsAxisError
		if !errors.As(err, &need) {
			return with, nil
		}
		if !interactive {
			return with, err
		}
		answer, askErr := askAxis(reader, out, need)
		if askErr != nil {
			return with, askErr
		}
		with = append(with, answer)
	}
	return with, nil
}

// maxAxisTries is how many times one axis is asked before giving up.
const maxAxisTries = 3

// askAxis puts the options of one axis, with the suggested one marked, and reads
// the answer: a number, a name, or an empty line for the suggested option.
//
// It returns an error when the input ends or after maxAxisTries bad answers.
func askAxis(in *bufio.Reader, out io.Writer, need *choiceselect.NeedsAxisError) (string, error) {
	if need.About != "" {
		fmt.Fprintf(out, "\n%s: %s\n", need.Choice, need.About)
	}
	fmt.Fprintf(out, "\n%s needs an answer on the %s axis. smith does not choose it for you:\n", need.Choice, need.Axis)
	suggested, hasSuggested := need.Suggested()
	for i, o := range need.Options {
		mark := ""
		if o.Suggested {
			mark = " (suggested)"
		}
		fmt.Fprintf(out, "  %d) %s%s: %s\n", i+1, o.Name, mark, o.Description)
	}
	for range maxAxisTries {
		if hasSuggested {
			fmt.Fprintf(out, "%s [%s]: ", need.Axis, suggested.Name)
		} else {
			fmt.Fprintf(out, "%s: ", need.Axis)
		}
		line, err := in.ReadString('\n')
		text := strings.TrimSpace(line)
		if err != nil && text == "" {
			return "", fmt.Errorf("no answer for the %s axis: the input ended (add one of %s to --with)", need.Axis, optionNames(need))
		}
		if text == "" && hasSuggested {
			return suggested.Name, nil
		}
		if n, convErr := strconv.Atoi(text); convErr == nil && n >= 1 && n <= len(need.Options) {
			return need.Options[n-1].Name, nil
		}
		for _, o := range need.Options {
			if o.Name == text {
				return o.Name, nil
			}
		}
		fmt.Fprintf(out, "%q is not one of the options: type a number or a name.\n", text)
	}
	return "", fmt.Errorf("no valid answer for the %s axis after %d tries (add one of %s to --with)", need.Axis, maxAxisTries, optionNames(need))
}

func optionNames(need *choiceselect.NeedsAxisError) string {
	names := make([]string, len(need.Options))
	for i, o := range need.Options {
		names[i] = o.Name
	}
	return strings.Join(names, ", ")
}
