package cachesetup

// The operations of the setup flow that the history records, one name each.
const (
	// StepChoose, StepText, StepConfirm and StepShow are questions: they ask
	// or show and change nothing, so the person may visit them again.
	StepChoose  = "choose"
	StepText    = "text"
	StepConfirm = "confirm"
	StepShow    = "show"
	// StepRun, StepWrite and StepSecret are actions: a program ran (az
	// listing, sign-in, the Azure variables), smith.toml was written, or a key
	// or token was made. They are points of no return: nothing undoes them.
	StepRun    = "run"
	StepWrite  = "write"
	StepSecret = "secret"
	// StepList is a program that only reads (a listing). It is not a question,
	// but it changes nothing, so going back passes over it and runs it again.
	StepList = "list"
)

// isQuestion lists, in one place, which steps are questions and which are
// actions. A step that is missing here counts as an action, the safe side.
var isQuestion = map[string]bool{
	StepChoose: true, StepText: true, StepConfirm: true, StepShow: true,
	StepRun: false, StepWrite: false, StepSecret: false, StepList: false,
}

// readsOnly lists the steps that are not questions but change nothing.
var readsOnly = map[string]bool{StepList: true}

// Steps are the operations the history records, with true for a question and
// false for an action (a point of no return). The result is a copy.
func Steps() map[string]bool {
	out := make(map[string]bool, len(isQuestion))
	for k, v := range isQuestion {
		out[k] = v
	}
	return out
}
