package recipechoice

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// maxRounds bounds the resolve-and-ask loop. Each round answers one choice,
// and a recipe chain declares far fewer choices than this.
const maxRounds = 64

// maxTries is how many times one question is asked before giving up.
const maxTries = 3

// Asker answers a recipe's choices.
type Asker struct {
	// In is read for answers, one line each; Out receives the questions.
	In  io.Reader
	Out io.Writer
	// Interactive says a person is there to answer. Without it every choice
	// not given takes its default, and one with no default is an error.
	Interactive bool
}

// Resolve resolves a recipe for the answers given and asks for the rest:
//
//   - interactively, every choice not given is asked once, in the order the
//     recipe declares them, a choice a when block adds only after the answer
//     that makes the block apply; the default is pre-selected, so an empty
//     line takes it
//   - otherwise a choice not given takes its default, and one with neither is
//     the resolver's MissingAnswerError, which names the choice and the
//     --choice flag that answers it
func (a Asker) Resolve(loaded recipe.Loaded, given map[string]string) (recipe.Resolution, error) {
	answers := map[string]string{}
	for name, answer := range given {
		answers[name] = answer
	}
	in := bufio.NewReader(a.In)
	for range maxRounds {
		res, err := recipe.Resolve(loaded, answers)
		var missing *recipe.MissingAnswerError
		switch {
		case errors.As(err, &missing):
			if !a.Interactive {
				return res, err
			}
			q := missing.Choices[0]
			answer, askErr := a.ask(in, q)
			if askErr != nil {
				return res, askErr
			}
			answers[q.Name] = answer
			continue
		case err != nil:
			return res, err
		}
		next, ok := firstUnanswered(res, answers)
		if !a.Interactive || !ok {
			return res, nil
		}
		answer, askErr := a.ask(in, next)
		if askErr != nil {
			return res, askErr
		}
		answers[next.Name] = answer
	}
	return recipe.Resolution{}, fmt.Errorf("the recipe asks more than %d choices", maxRounds)
}

// firstUnanswered is the first choice the resolution asked that the caller
// has not answered yet (it took its default).
func firstUnanswered(res recipe.Resolution, answers map[string]string) (recipe.Question, bool) {
	for _, q := range res.Asked {
		if _, done := answers[q.Name]; !done {
			return q, true
		}
	}
	return recipe.Question{}, false
}

// ask puts one question and reads its answer: an option's text or number, the
// default on an empty line, or any text for a choice without options.
func (a Asker) ask(in *bufio.Reader, q recipe.Question) (string, error) {
	label := q.Ask
	if label == "" {
		label = q.Name
	}
	for try := 0; try < maxTries; try++ {
		fmt.Fprint(a.Out, prompt(label, q))
		line, err := in.ReadString('\n')
		text := strings.TrimSpace(line)
		if err != nil && text == "" {
			return "", fmt.Errorf("no answer for the choice %q: the input ended (give one with --choice %s=<answer>)", q.Name, q.Name)
		}
		answer, problem := interpret(text, q)
		if problem == "" {
			return answer, nil
		}
		fmt.Fprintln(a.Out, problem)
	}
	return "", fmt.Errorf("no valid answer for the choice %q after %d tries", q.Name, maxTries)
}

func prompt(label string, q recipe.Question) string {
	var b strings.Builder
	b.WriteString(label)
	if len(q.Options) > 0 {
		b.WriteString(" (")
		for i, o := range q.Options {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%d %s", i+1, o)
		}
		b.WriteString(")")
	}
	if q.Default != "" {
		fmt.Fprintf(&b, " [%s]", q.Default)
	}
	b.WriteString(": ")
	return b.String()
}

// interpret turns a typed line into an answer, or says what is wrong with it.
func interpret(text string, q recipe.Question) (answer, problem string) {
	if text == "" {
		if q.Default != "" {
			return q.Default, ""
		}
		return "", "an answer is required"
	}
	if len(q.Options) == 0 {
		return text, ""
	}
	for _, o := range q.Options {
		if o == text {
			return o, ""
		}
	}
	if n, err := strconv.Atoi(text); err == nil && n >= 1 && n <= len(q.Options) {
		return q.Options[n-1], ""
	}
	return "", fmt.Sprintf("choose one of: %s", strings.Join(q.Options, ", "))
}
