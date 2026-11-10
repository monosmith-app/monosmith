package cachepipeline

import "strings"

// Failure is why an az command could not do its work.
type Failure int

// The causes smith knows how to explain.
const (
	// FailureOther is any other cause; the error text of az tells.
	FailureOther Failure = iota
	// FailureSignIn: nobody is signed in to az.
	FailureSignIn
	// FailureExtension: az has no azure-devops extension.
	FailureExtension
)

// Classify reads the error text of a failed az command.
//
// Parameters: stderr is what az wrote to its error output.
// Returns: the cause, FailureOther when it is not one smith knows.
func Classify(stderr string) Failure {
	text := strings.ToLower(stderr)
	for _, m := range []string{"az login", "az devops login", "aadsts", "not logged in", "vs30063", "tf400813"} {
		if strings.Contains(text, m) {
			return FailureSignIn
		}
	}
	for _, m := range []string{"misspelled or not recognized", "az extension add", "azure-devops"} {
		if strings.Contains(text, m) {
			return FailureExtension
		}
	}
	return FailureOther
}
