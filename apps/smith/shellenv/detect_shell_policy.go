package shellenv

// DetectShell finds the shell the person works in.
//
// The order is: the program that started smith (when it is a shell), the
// PowerShell variable PSModulePath (outside Windows, where only PowerShell
// sets it), the SHELL variable, and on Windows PowerShell. The parent program
// comes first because it is the shell the person typed the command in. It is
// not a shell when an npm launcher started smith, so the search goes on.
//
// Parameters: env is the machine.
// Returns: the shell, and false when no rule found one.
func DetectShell(env Environment) (Shell, bool) {
	if s, ok := ParseShell(env.ParentProgram()); ok {
		return s, true
	}
	if env.OS() != "windows" && env.Getenv("PSModulePath") != "" {
		return PowerShell, true
	}
	if s, ok := ParseShell(env.Getenv("SHELL")); ok {
		return s, true
	}
	if env.OS() == "windows" {
		return PowerShell, true
	}
	return "", false
}
