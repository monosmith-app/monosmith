package tauriadd

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

var (
	identifierField = regexp.MustCompile(`("identifier"\s*:\s*)"[^"]*"`)
	identifierShape = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*(\.[A-Za-z][A-Za-z0-9-]*)+$`)
)

// DefaultIdentifier is the bundle identifier smith proposes: com.<workspace>.<project>,
// each part lowered to letters, digits and hyphens.
//
// Parameters: workspace is the workspace name; project is the app's name.
// Returns: the proposed identifier.
func DefaultIdentifier(workspace, project string) string {
	return "com." + segment(workspace) + "." + segment(project)
}

func segment(s string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(s) {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' {
			b.WriteRune(c)
		}
	}
	if b.Len() == 0 {
		return "app"
	}
	return b.String()
}

// ValidIdentifier checks a bundle identifier: two or more dot-separated parts,
// each starting with a letter and holding letters, digits and hyphens.
//
// Parameters: id is the identifier.
// Returns: nil when valid; a problem that shows the shape when not.
func ValidIdentifier(id string) error {
	if identifierShape.MatchString(id) {
		return nil
	}
	return problem.New("\""+id+"\" is not a bundle identifier",
		"write it like com.yourcompany.yourapp: parts joined by dots, each starting with a letter, with letters, digits and hyphens")
}

// SetIdentifier replaces the identifier in the text of a tauri.conf.json.
//
// Parameters: conf is the file's text; id is the identifier to write.
// Returns: the new text, and false when the file has no identifier field.
func SetIdentifier(conf, id string) (string, bool) {
	if !identifierField.MatchString(conf) {
		return conf, false
	}
	return identifierField.ReplaceAllString(conf, `${1}"`+id+`"`), true
}
