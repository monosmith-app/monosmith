package recipechoice

import (
	"fmt"
	"strings"
)

// ParseAnswers turns repeated --choice name=answer flags into the answers
// the resolver takes. A choice given twice, or a flag with no name or no
// answer, is an error.
func ParseAnswers(pairs []string) (map[string]string, error) {
	answers := map[string]string{}
	for _, pair := range pairs {
		name, answer, ok := strings.Cut(pair, "=")
		name, answer = strings.TrimSpace(name), strings.TrimSpace(answer)
		if !ok || name == "" || answer == "" {
			return nil, fmt.Errorf("--choice %q: write it as name=answer", pair)
		}
		if previous, dup := answers[name]; dup {
			return nil, fmt.Errorf("--choice %s is given twice (%s and %s)", name, previous, answer)
		}
		answers[name] = answer
	}
	return answers, nil
}
