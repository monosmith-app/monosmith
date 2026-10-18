package recipe

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// maxScenarios bounds how many combinations of answers ValidateAll resolves.
const maxScenarios = 64

// ValidateAll resolves and validates a recipe for every combination of
// answers to the choices it asks (at most 64 combinations), so a branch that
// is fine for python and broken for csharp is found, not only the branch the
// defaults pick. It stops at the first problem and names the answers that
// lead to it.
func ValidateAll(loaded Loaded, lookup PluginLookup) error {
	seen := map[string]bool{}
	queue := []map[string]string{{}}
	for n := 0; len(queue) > 0 && n < maxScenarios; {
		answers := queue[0]
		queue = queue[1:]
		key := answersKey(answers)
		if seen[key] {
			continue
		}
		seen[key] = true
		n++
		res, err := Resolve(loaded, answers)
		var missing *MissingAnswerError
		if errors.As(err, &missing) {
			q := missing.Choices[0]
			if len(q.Options) == 0 {
				return fmt.Errorf("choice %q (%s) has no options and no default, so it cannot be checked", q.Name, q.Origin)
			}
			for _, option := range q.Options {
				queue = append(queue, with(answers, q.Name, option))
			}
			continue
		}
		if err == nil {
			err = Validate(res, lookup)
		}
		if err != nil {
			return forAnswers(answers, err)
		}
		for _, q := range res.Asked {
			if _, given := answers[q.Name]; given {
				continue
			}
			for _, option := range q.Options {
				if option != q.Answer {
					queue = append(queue, with(answers, q.Name, option))
				}
			}
		}
	}
	return nil
}

func with(answers map[string]string, name, answer string) map[string]string {
	out := make(map[string]string, len(answers)+1)
	for k, v := range answers {
		out[k] = v
	}
	out[name] = answer
	return out
}

// answersKey names a combination of answers, the same whatever the order.
func answersKey(answers map[string]string) string {
	keys := make([]string, 0, len(answers))
	for k, v := range answers {
		keys = append(keys, k+"="+v)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// forAnswers says which answers an error belongs to.
func forAnswers(answers map[string]string, err error) error {
	if len(answers) == 0 {
		return err
	}
	prefix := "with " + strings.ReplaceAll(answersKey(answers), ",", ", ") + ": "
	var invalid *ValidationError
	if errors.As(err, &invalid) {
		return &ValidationError{Message: prefix + invalid.Message, Chosen: invalid.Chosen}
	}
	return fmt.Errorf("%s%w", prefix, err)
}
