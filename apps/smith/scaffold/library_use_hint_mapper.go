package scaffold

// libraryUseHint is the line smith add prints after it made a library: how an
// app of the workspace uses it. The library is not an app's dependency until
// someone says so, and a new person does not know that command.
//
// Parameters: kind is the kind that was added (app, lib or package); name is the project.
// Returns: the line, or "" for a kind that other projects do not import.
func libraryUseHint(kind, name string) string {
	if kind != "lib" {
		return ""
	}
	return "Use it from an app: smith install -w <app> " + name + " (in the terminal interface, press + on the app and pick it)."
}
