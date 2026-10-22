package screenreader

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

// maxTries is how many times one question is asked before giving up.
const maxTries = 3

// Ask walks the guided actions as plain lines: a numbered list to choose from,
// then each of the action's questions on a line of its own, answered by typing
// a number (or, for a free question, the text) and Enter. Nothing is drawn
// over earlier output, no colour or cursor movement is used, and every line
// reads on its own, so a screen reader announces the conversation as it goes.
// It returns the smith arguments the answers build, or nil when the user quit
// (q, or the input ended).
func Ask(in io.Reader, out io.Writer, actions []guide.Action) ([]string, error) {
	r := bufio.NewReader(in)
	if len(actions) == 0 {
		fmt.Fprintln(out, "There is nothing to do here.")
		return nil, nil
	}
	fmt.Fprintln(out, "smith. What do you want to do? Type a number and press Enter, or q to quit.")
	for i, a := range actions {
		fmt.Fprintf(out, "%d. %s\n", i+1, describe(a.Title, a.Hint))
	}
	for try := 0; try < maxTries; try++ {
		fmt.Fprint(out, "Your choice: ")
		line, quit := readLine(r)
		if quit {
			return nil, nil
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(actions) {
			fmt.Fprintf(out, "%q is not one of the numbers 1 to %d.\n", line, len(actions))
			continue
		}
		return ask(r, out, actions[n-1])
	}
	return nil, fmt.Errorf("no valid choice after %d tries", maxTries)
}

// ask puts an action's questions and builds its command line.
func ask(r *bufio.Reader, out io.Writer, a guide.Action) ([]string, error) {
	answers := map[string][]string{}
	for _, step := range a.Steps {
		got, quit, err := question(r, out, step)
		if err != nil {
			return nil, err
		}
		if quit {
			return nil, nil
		}
		answers[step.Key] = got
	}
	if a.Argv == nil {
		return nil, nil
	}
	return a.Argv(answers), nil
}

// question asks one step until it has a valid answer.
func question(r *bufio.Reader, out io.Writer, step guide.Step) (answer []string, quit bool, err error) {
	if len(step.Options) == 0 {
		hint := ""
		if step.Placeholder != "" {
			hint = " For example " + step.Placeholder + "."
		}
		if step.Optional {
			hint += " Press Enter to skip."
		}
		fmt.Fprintf(out, "%s.%s\n", strings.TrimSuffix(step.Prompt, "."), hint)
		for try := 0; try < maxTries; try++ {
			fmt.Fprint(out, "Your answer: ")
			line, q := readLine(r)
			if q {
				return nil, true, nil
			}
			if line != "" || step.Optional {
				return []string{line}, false, nil
			}
			fmt.Fprintln(out, "An answer is required.")
		}
		return nil, false, fmt.Errorf("no answer to %q after %d tries", step.Prompt, maxTries)
	}
	how := "Type the number of one."
	if step.Multi {
		how = "Type the numbers of the ones you want, separated by spaces."
	}
	fmt.Fprintf(out, "%s %s\n", strings.TrimSuffix(step.Prompt, "."), how)
	for i, o := range step.Options {
		fmt.Fprintf(out, "%d. %s\n", i+1, describe(o.Label, o.Hint))
	}
	for try := 0; try < maxTries; try++ {
		fmt.Fprint(out, "Your answer: ")
		line, q := readLine(r)
		if q {
			return nil, true, nil
		}
		values, problem := pick(line, step)
		if problem == "" {
			return values, false, nil
		}
		fmt.Fprintln(out, problem)
	}
	return nil, false, fmt.Errorf("no valid answer to %q after %d tries", step.Prompt, maxTries)
}

// pick turns the typed numbers into the options' values, or says what is wrong.
func pick(line string, step guide.Step) (values []string, problem string) {
	fields := strings.Fields(strings.ReplaceAll(line, ",", " "))
	if len(fields) == 0 {
		return nil, "Type at least one number."
	}
	if !step.Multi && len(fields) > 1 {
		return nil, "Type one number."
	}
	for _, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil || n < 1 || n > len(step.Options) {
			return nil, fmt.Sprintf("%q is not one of the numbers 1 to %d.", f, len(step.Options))
		}
		values = append(values, step.Options[n-1].Value)
	}
	return values, ""
}

// describe is a title with its hint after a dash.
func describe(title, hint string) string {
	if hint == "" {
		return title
	}
	return title + " - " + hint
}

// readLine reads one trimmed line; quit is true for q or when the input ended.
func readLine(r *bufio.Reader) (line string, quit bool) {
	text, err := r.ReadString('\n')
	text = strings.TrimSpace(text)
	if err != nil && text == "" {
		return "", true
	}
	if strings.EqualFold(text, "q") || strings.EqualFold(text, "quit") {
		return "", true
	}
	return text, false
}
