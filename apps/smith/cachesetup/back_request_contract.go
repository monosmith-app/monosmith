package cachesetup

// BackRequest is what a Dialogue panics with when the person asks to go back
// to the question before the one on screen. Run catches it: it asks the
// earlier question again, or says why it cannot (see Steps). No other code
// recovers it.
type BackRequest struct{}

// BackAware is implemented by a Dialogue that offers a way back. Run calls
// AllowBack before each question: can is true when a question before this one
// exists, so the first question shows no way back and keeps its old meaning
// of Esc (close).
type BackAware interface {
	// AllowBack tells the dialogue whether the next question may be left with Back.
	AllowBack(can bool)
}
