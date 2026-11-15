package release

import (
	"regexp"
	"strings"
)

var (
	headerPattern    = regexp.MustCompile(`^([A-Za-z]+)(?:\(([^)]*)\))?(!)?:\s+(.+)$`)
	breakingFooter   = regexp.MustCompile(`(?m)^BREAKING[ -]CHANGE:\s`)
	directivePattern = regexp.MustCompile(`^version\(([^)]+)\)\[([^\]]+)\]:`)
)

// Commit is a parsed conventional commit.
type Commit struct {
	Hash     string
	Type     string
	Scopes   []string
	Breaking bool
	// BreakingNote is the BREAKING CHANGE footer's text, up to a blank line.
	BreakingNote string
	Description  string
	Subject      string
	// Valid is false for a subject that is not a conventional commit.
	Valid bool
}

// ParseCommit parses a commit's subject and body.
func ParseCommit(hash, subject, body string) Commit {
	c := Commit{Hash: hash, Subject: subject}
	m := headerPattern.FindStringSubmatch(strings.TrimSpace(subject))
	if m == nil {
		return c
	}
	c.Valid = true
	c.Type = strings.ToLower(m[1])
	for _, s := range strings.Split(m[2], ",") {
		if s = strings.TrimSpace(s); s != "" {
			c.Scopes = append(c.Scopes, s)
		}
	}
	c.Breaking = m[3] == "!" || breakingFooter.MatchString(body)
	if loc := breakingFooter.FindStringIndex(body); loc != nil {
		note, _, _ := strings.Cut(strings.ReplaceAll(body[loc[1]:], "\r\n", "\n"), "\n\n")
		c.BreakingNote = strings.TrimSpace(note)
	}
	c.Description = m[4]
	return c
}

// Directive is a commit that forces a project's next version:
// "version(web)[2.0.0]: start the 2.x line".
type Directive struct {
	Project, Version string
}

// ParseDirective returns the directive in a subject, if any.
func ParseDirective(subject string) (Directive, bool) {
	m := directivePattern.FindStringSubmatch(strings.TrimSpace(subject))
	if m == nil {
		return Directive{}, false
	}
	return Directive{Project: strings.TrimSpace(m[1]), Version: strings.TrimSpace(m[2])}, true
}

// IsReleaseCommit reports commits made by a release itself, which never cause
// another release.
func IsReleaseCommit(c Commit) bool {
	return c.Type == "chore" && len(c.Scopes) == 1 && c.Scopes[0] == "release"
}
