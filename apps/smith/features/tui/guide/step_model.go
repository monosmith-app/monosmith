package guide

// Step is one question: a choice from options, or free text when Options is nil.
type Step struct {
	Key     string
	Prompt  string
	Options []Option
	// OptionsFor, when set, makes the options from the earlier answers (the names
	// typed in a question before): it replaces Options.
	OptionsFor func(answers map[string][]string) []Option
	// Take, when set, answers the question for the person when it says ok (the
	// defaults of a form were chosen): the question is then not asked.
	Take func(answers map[string][]string) (answer []string, ok bool)
	// Placeholder is shown in an empty text field.
	Placeholder string
	// Multi lets the user toggle several options (space) before confirming.
	Multi bool
	// Optional lets a free-text step be left empty; the answer is then empty.
	Optional bool
	// Filter lets typing narrow the options of a long list.
	Filter bool
	// PlaceholderFor, when set, picks the placeholder from the earlier answers.
	PlaceholderFor func(answers map[string][]string) string
	// Check says what is wrong with a typed answer ("" when it is fine); it
	// runs as the person types and again on enter.
	Check func(answers map[string][]string, value string) string
	// Suggest, when set, offers a corrected answer for a typed one that Check
	// refuses ("" when it has none). Tab takes it; typing on keeps the choice
	// to write another.
	Suggest func(answers map[string][]string, value string) string
	// Skip, when set, leaves the question out when it says so, given the
	// earlier answers; the key then has no answer.
	Skip func(answers map[string][]string) bool
	// Initial, when set, fills a free-text step with a value the person edits
	// (the names an app's environments start on). The grey placeholder is only
	// an example; this is an answer that Enter takes.
	Initial func(answers map[string][]string) string
	// Auto, when set, answers the question itself when it returns a value (the
	// answer is then the same as a pick of that option) and asks it when it
	// returns "": a language that an outcome has only one of is not asked.
	Auto func(answers map[string][]string) string
	// Ticked, when set, names the options of a multiple choice that start
	// ticked, given the earlier answers.
	Ticked func(answers map[string][]string) []string
	// Preselect, when set, picks the option the cursor starts on from the words
	// typed in the last list that was filtered ("electron" starts on Electron).
	Preselect func(typed string, options []Option) string
	// Note is a muted line under the question; Try is the examples shown in the
	// place of "Type to narrow the list" on a Filter step.
	Note, Try string
	// Glossary marks a Note that explains the words of the list: on a short
	// screen it folds to one line, and "?" opens it.
	Glossary bool
	// Path makes the free-text step ask for a path: the list of folders and
	// files of the folder being typed shows under the field (package
	// pathfinder). What is typed is still the answer; the list only helps.
	Path bool
	// PathSettings also offers the settings of smith.toml (ci.devcontainer) in
	// that list, for a question that takes a path or a setting.
	PathSettings bool
}

// Option is one answer.
type Option struct {
	Label, Value, Hint string
	// Group names the single-choice axis the option sits on ("framework"); a
	// multiple choice then ticks one option of a group at a time. Empty means
	// the option can be ticked next to any other. Many lets the option share
	// its group with other options that are Many too.
	Group string
	Many  bool
	// Also are other names that find the option when a list is filtered: the
	// frameworks, test runners and styles a language offers ("react" finds Node).
	Also []string
	// Wizard, when set, ends the form on this option and hands the person to
	// the wizard or form it names (WizardRecipe), as Action.Wizard does.
	Wizard string
}

// Action is a menu entry: its questions, then the command they build.
type Action struct {
	Title, Hint string
	Steps       []Step
	// Argv turns the answers into smith's arguments.
	Argv func(answers map[string][]string) []string
	// NeedsWorkspace hides the action outside a workspace.
	NeedsWorkspace bool
	// OutsideOnly hides the action inside a workspace.
	OutsideOnly bool
	// NeedsContent hides the action in a folder that holds nothing.
	NeedsContent bool
	// Presents is the id of the action (see package actions) the form runs.
	Presents string
	// Wizard, when set, hands the action to the shell's wizard of that name
	// instead of asking its Steps: the plain, screen-reader menu still asks
	// them.
	Wizard string
}
