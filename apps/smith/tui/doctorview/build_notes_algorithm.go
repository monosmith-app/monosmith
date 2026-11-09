package doctorview

import (
	"regexp"
	"strings"
)

// buildNotes match what a tool prints about how it was built: the Go version,
// the module origin with its checksum, the platform. They tell a maintainer
// nothing about what is wrong, so the panel folds them away.
var buildNotes = []*regexp.Regexp{
	regexp.MustCompile(` built with go[0-9][\w.]*`),                                // words:literal
	regexp.MustCompile(` from \([^()]*(?:unknown|modified|mod sum|h1:)[^()]*\)`),   // words:literal
	regexp.MustCompile(` on (?:\(unknown\)|(?:windows|linux|darwin|freebsd)/\w+)`), // words:literal
	regexp.MustCompile(`,? ?(?:mod )?sum: ?"?h1:[A-Za-z0-9+/=]+"?`),                // words:literal
	regexp.MustCompile(`h1:[A-Za-z0-9+/=]{16,}`),                                   // words:literal
}

// SplitBuildNotes takes the build metadata out of one sentence of a check.
//
// Parameters: sentence is one problem of a check, as a plugin wrote it.
// Returns: cause is the sentence without the metadata, on one line; notes is
// the metadata that was cut, in the order it was found ("" when there is none).
func SplitBuildNotes(sentence string) (cause, notes string) {
	cause, _, _ = strings.Cut(sentence, "\n")
	var found []string
	for _, re := range buildNotes {
		for _, m := range re.FindAllString(cause, -1) {
			found = append(found, strings.TrimSpace(strings.TrimLeft(m, ", ")))
		}
		cause = re.ReplaceAllString(cause, "")
	}
	cause = strings.Join(strings.Fields(cause), " ")
	return strings.TrimSpace(cause), strings.Join(found, " ")
}
