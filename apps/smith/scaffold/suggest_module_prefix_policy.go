package scaffold

// SuggestModulePrefix is the Go module prefix smith proposes for a workspace
// that has no git remote: github.com/<git user>/<name> when git knows the
// user, else the placeholder example.com/<name>.
//
// Parameters: name is the workspace name.
// Returns: the prefix, and true when it comes from the git user (and is a
// guess to check) rather than the placeholder.
func SuggestModulePrefix(name string) (string, bool) {
	if user := gitUserSlug(); user != "" {
		return "github.com/" + user + "/" + name, true
	}
	return "example.com/" + name, false
}
