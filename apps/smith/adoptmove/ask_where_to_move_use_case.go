package adoptmove

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// maxTries is how many bad answers AskWhere accepts before it gives up.
const maxTries = 3

// AskWhere prints the explanation and the numbered choices and reads the
// answer from in: a number (an empty line takes the suggested one), then the
// name (an empty line keeps the derived one) or, for "another folder", the
// folder. It returns the folder, or Leave. It gives up after three bad answers.
// Nothing is changed.
func AskWhere(c Candidate, in io.Reader, out io.Writer) (string, error) {
	r := bufio.NewReader(in)
	choices := Choices(c)
	suggested := DefaultChoice(choices)
	fmt.Fprintf(out, "%s\n\n%s", Explanation(c), QuestionText(c))
	ask := func(prompt string) (string, error) {
		fmt.Fprint(out, prompt)
		line, err := r.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			return "", problem.New("no answer was given", Command("<folder>")+" says it without a question")
		}
		// A byte order mark can lead piped input (PowerShell writes one).
		return strings.TrimSpace(strings.TrimPrefix(line, string(rune(0xFEFF)))), nil
	}
	for try := 0; try < maxTries; try++ {
		answer, err := ask(fmt.Sprintf("Choose 1-%d [%d]: ", len(choices), suggested))
		if err != nil {
			return "", err
		}
		n := suggested
		if answer != "" {
			var perr error
			if n, perr = strconv.Atoi(answer); perr != nil || n < 1 || n > len(choices) {
				fmt.Fprintf(out, "  %q is not a choice; type a number from 1 to %d.\n", answer, len(choices))
				continue
			}
		}
		choice := choices[n-1]
		if choice.Key == KeyNone {
			return Leave, nil
		}
		prompt, fallback := fmt.Sprintf("Name [%s]: ", c.Name), c.Name
		if choice.Key == KeyOther {
			prompt, fallback = "Folder, from the repository root (for example tools/"+c.Name+"): ", ""
		}
		typed, err := ask(prompt)
		if err != nil {
			return "", err
		}
		if typed == "" {
			typed = fallback
		}
		dir, err := Destination(choice.Key, typed)
		if err != nil {
			fmt.Fprintln(out, "  "+err.Error())
			continue
		}
		return dir, nil
	}
	return "", problem.New("no valid answer after three tries; nothing was changed", Command("<folder>"))
}
