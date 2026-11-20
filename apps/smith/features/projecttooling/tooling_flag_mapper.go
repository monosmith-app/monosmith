package projecttooling

import (
	"fmt"
	"strings"
)

// Answers are the answers given on the command line: one for every project,
// one for each named project, or none.
type Answers struct {
	// All answers every project that has no answer of its own; "" is none.
	All Answer
	// Per maps a project name to its answer.
	Per map[string]Answer
}

// For is the answer given for the project, and false when none was given.
//
// Parameters: name is the project's name.
// Returns: the answer and whether one was given.
func (a Answers) For(name string) (Answer, bool) {
	if v, ok := a.Per[name]; ok {
		return v, true
	}
	return a.All, a.All != ""
}

// ParseAnswers reads the values of --tooling: "smith" or "keep" answers every
// project, "<project>=smith" or "<project>=keep" answers one.
//
// Parameters: values are the flag's values, in order.
// Returns: the answers; an error names a value that is not one of these forms.
func ParseAnswers(values []string) (Answers, error) {
	var a Answers
	for _, v := range values {
		name, answer, named := strings.Cut(strings.TrimSpace(v), "=")
		if !named {
			answer, name = name, ""
		}
		parsed, err := parseAnswer(answer)
		if err != nil {
			return Answers{}, fmt.Errorf("--tooling %q: %w", v, err)
		}
		if !named {
			a.All = parsed
			continue
		}
		if name == "" {
			return Answers{}, fmt.Errorf("--tooling %q: the project name before = is empty", v)
		}
		if a.Per == nil {
			a.Per = map[string]Answer{}
		}
		a.Per[name] = parsed
	}
	return a, nil
}

func parseAnswer(s string) (Answer, error) {
	switch Answer(s) {
	case Keep, Smith:
		return Answer(s), nil
	}
	return "", fmt.Errorf("%q is not smith or keep", s)
}
