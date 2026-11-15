package recipe

import (
	"fmt"
	"strings"
)

// ParseError is a recipe file that could not be read; its message names the
// file and line.
type ParseError struct {
	Source  string
	Message string
}

func (e *ParseError) Error() string { return e.Message }

// ChangedError is a URL recipe whose content is not what the lockfile pinned.
// The upgrade command accepts the new content (Loader.Accept).
type ChangedError struct {
	URL    string
	Locked string
	Got    string
}

func (e *ChangedError) Error() string {
	return fmt.Sprintf("the recipe at %s changed since it was pinned (locked sha256 %s, now %s); review it, then run `smith upgrade` to accept the new content", e.URL, short(e.Locked), short(e.Got))
}

func short(sum string) string {
	if len(sum) > 12 {
		return sum[:12]
	}
	return sum
}

// MissingAnswerError lists the choices that have no answer and no default.
type MissingAnswerError struct {
	Choices []Question
}

func (e *MissingAnswerError) Error() string {
	var parts []string
	for _, q := range e.Choices {
		hint := "--choice " + q.Name + "=<answer>"
		if len(q.Options) > 0 {
			hint = "--choice " + q.Name + "=" + strings.Join(q.Options, "|")
		}
		parts = append(parts, fmt.Sprintf("choice %q (%s) has no answer and no default; give one with %s", q.Name, q.Origin, hint))
	}
	return strings.Join(parts, "\n")
}

// ValidationError is a resolved recipe whose plugins cannot work together. It
// carries where each plugin named in the message was chosen.
type ValidationError struct {
	Message string
	// Chosen are the recipe lines that chose the plugins the message names.
	Chosen []string
}

func (e *ValidationError) Error() string {
	if len(e.Chosen) == 0 {
		return e.Message
	}
	return e.Message + "\n  chosen at:\n    " + strings.Join(e.Chosen, "\n    ")
}
