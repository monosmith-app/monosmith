package tui

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// oncePerThing are the commands that make or remove something: the same line
// run a second time fails or does nothing, because the thing exists already or
// is gone. They are not offered again.
var oncePerThing = []string{"add", "new", "remove", "adopt", "convert"}

// repeatable reports whether a command line makes sense to run again as it was.
//
// Parameters: argv is the smith command line; nil is not repeatable.
// Returns: false for a command that makes or removes something.
func repeatable(argv []string) bool {
	return len(argv) > 0 && !slices.Contains(oncePerThing, argv[0])
}

// notRepeatableReason says why a command is not run again and what to do
// instead, for the key and the action that would have run it.
func notRepeatableReason(argv []string) string {
	verb := words.Say("tui.repeat.that_command")
	if len(argv) > 0 {
		verb = words.Say("tui.repeat.smith", "argv", argv[0])
	}
	return words.Say("tui.repeat.makes_or_removes_something_so_running", "verb", verb)
}
