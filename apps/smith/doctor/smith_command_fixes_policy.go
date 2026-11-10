package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// deriveChoice is the fix of a derived file that someone edited: the person
// decides who owns the file. The first answer is the suggested one.
//
// Parameters: path is the file.
// Returns: the fix, which asks one question.
func deriveChoice(path string) *protocol.Fix {
	return smithfix.Choice("Settle the derived file "+path, "Who owns "+path+" from now on? You see the difference before anything is written.",
		smithfix.Option("smith: overwrite the file with what smith derives (your edit is lost)", "derive", "--replace", path),
		smithfix.Option("You, as the new baseline: keep your edit and let smith follow it (keys-mode files)", "derive", "--adopt", path),
		smithfix.Option("You, for good: smith stops touching this file", "derive", "--eject", path))
}

// cacheFix is the fix of a cache check that failed.
//
// Parameters: ok is true when the check passed, so there is nothing to fix;
// description names the fix; args is the smith command line.
// Returns: the fix, or nil when ok.
func cacheFix(ok bool, description string, args ...string) *protocol.Fix {
	if ok {
		return nil
	}
	return smithfix.Fix(description, smithfix.Step(description, args...))
}
