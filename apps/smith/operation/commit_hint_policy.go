package operation

import "os"

// InterfaceEnv names the environment variable that tells a smith command which
// interface started it. The terminal interface sets it to InterfaceTUI, so the
// command words its advice for a person who has no flags to type.
const InterfaceEnv = "SMITH_INTERFACE"

// InterfaceTUI is the value of InterfaceEnv inside the terminal interface.
const InterfaceTUI = "tui"

// CommitNextTime is the words that tell the person how to make the next change
// a commit, to put before "next time": the flag at a command line, the box in
// the terminal interface.
//
// Parameters: atCommandLine is the advice for a shell user ("add --commit").
// Returns: atCommandLine, or `tick "Commit this change"` inside the terminal
// interface.
func CommitNextTime(atCommandLine string) string {
	if os.Getenv(InterfaceEnv) == InterfaceTUI {
		return `tick "Commit this change"`
	}
	return atCommandLine
}
