package choiceselect

// IncompleteError says a selection is not wrong but not finished: no choice
// writes the project yet, a required choice is not in it, or a choice has no
// plugin to build on. A person who composes a selection step by step sees it as
// the next thing to add, not as a conflict. Its text is the sentence the
// solver always gave.
type IncompleteError struct {
	// Message is the whole sentence: the problem, then what to add.
	Message string
}

// Error returns the sentence.
func (e *IncompleteError) Error() string { return e.Message }
