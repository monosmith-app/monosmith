package fixwizard

// Style words the dialog for a walk that is not a doctor fix. The zero Style
// is the fix wizard itself, so a caller sets only what differs.
type Style struct {
	// Name starts the titles: "Fix · review" becomes "Adopt · review".
	Name string
	// Review says what the review shows, in place of "review": "Bring in ·
	// what smith found". Ask does the same for a question ("question").
	Review, Ask string
	// Verb is the title while a step runs ("Fixing…").
	Verb string
	// Working says what happens before the first step reports ("writing the files").
	Working string
	// Apply labels the button that confirms the review ("Apply").
	Apply string
	// Decline and Skip, when set, add two answers to the review: n declines
	// this one (Outcome.Decline), s declines it and everything after it
	// (Outcome.Skip). Without them n cancels.
	Decline, Skip string
	// Back adds a Back button to the first question and to the review, even
	// when no question was answered: it asks the caller to go to the stage
	// before (Outcome.Back).
	Back bool
	// NoRetry leaves Retry out of a failure: what failed cannot be run again.
	NoRetry bool
	// Input, when set, adds one editable text to every question, named by it
	// ("Name"). e starts the editing, Enter or Esc ends it, and "{name}" in an
	// option's label shows the text as it is typed. The caller reads it with
	// Wizard.Text.
	Input string
}

func (s Style) name() string {
	if s.Name == "" {
		return "Fix"
	}
	return s.Name
}

func (s Style) review() string {
	if s.Review == "" {
		return "review"
	}
	return s.Review
}

func (s Style) ask() string {
	if s.Ask == "" {
		return "question"
	}
	return s.Ask
}

func (s Style) verb() string {
	if s.Verb == "" {
		return "Fixing…"
	}
	return s.Verb
}

func (s Style) working() string {
	if s.Working == "" {
		return "writing the files"
	}
	return s.Working
}

func (s Style) apply() string {
	if s.Apply == "" {
		return "Apply"
	}
	return s.Apply
}
