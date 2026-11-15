package sandbox

// Runner runs the smith command line args in dir and returns its combined
// output and exit code. env is added to the process environment. The real
// runner starts this smith's own binary; tests pass a fake.
type Runner func(dir string, env []string, args ...string) (output string, code int, err error)

// Allowed are the commands `smith try` accepts: the ones that change files.
var Allowed = []string{"add", "remove", "up", "upgrade", "derive", "kind", "eject", "sync", "bootstrap"}

// Checks are the names a try can run after the command.
var Checks = []string{"doctor", "build", "test", "lint", "typecheck"}

// CheckResult is one check run in the copy.
type CheckResult struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
	// Tail is the last lines of its output, kept for a failure.
	Tail string `json:"tail,omitempty"`
}

// Result is what happened in the copy.
type Result struct {
	// Command is the command line that ran.
	Command []string `json:"command"`
	// CommandOK is false when the command itself failed.
	CommandOK bool `json:"commandOk"`
	// CommandTail is the last lines of the command's output.
	CommandTail string `json:"commandTail,omitempty"`
	// Changed are the files the command changed in the copy, as git status
	// writes them ("M  path").
	Changed []string      `json:"changed"`
	Checks  []CheckResult `json:"checks"`
	// Kept is the copy's folder when it was kept.
	Kept string `json:"kept,omitempty"`
}

// OK reports whether the command and every check passed.
func (r Result) OK() bool {
	if !r.CommandOK {
		return false
	}
	for _, c := range r.Checks {
		if !c.OK {
			return false
		}
	}
	return true
}
