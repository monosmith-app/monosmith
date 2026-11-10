package operationplan

// Environment says who is at the other end of the command.
type Environment struct {
	// Terminal is true when stdin and stdout are terminals.
	Terminal bool
	// CI is true when the CI environment variable (or a CI provider's) is set.
	CI bool
}

// ShouldAsk reports whether a command asks before it writes: a person is at a
// terminal, this is not CI, and --yes was not given. Every other run proceeds
// without a question, so a script never blocks.
//
// Parameters: env is the terminal state; yes is the value of --yes.
func ShouldAsk(env Environment, yes bool) bool {
	return env.Terminal && !env.CI && !yes
}
