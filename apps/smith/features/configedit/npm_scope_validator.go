package configedit

import (
	"fmt"
	"regexp"
)

// npmScope is what npm accepts as a scope: an @ and a lowercase name.
var npmScope = regexp.MustCompile(`^@[a-z0-9-~][a-z0-9-._~]*$`)

// checkNpmScope refuses a value that is not an npm scope for the field at
// path. Only workspace.scope holds a scope, and an empty value means "none".
//
// Parameters: path is the setting being edited; value is the text to check.
// Returns: an error that says what a scope looks like, or nil.
func checkNpmScope(path, value string) error {
	if path == "workspace.scope" && value != "" && !npmScope.MatchString(value) {
		return fmt.Errorf("%s is an npm scope: an @ and lowercase words joined by dashes, like @acme, not %q", path, value)
	}
	return nil
}
