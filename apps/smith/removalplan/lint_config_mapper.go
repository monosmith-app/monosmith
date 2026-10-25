package removalplan

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

// LintConfigPackage is the shared ESLint config smith uses, and the range
// that installs it.
const (
	LintConfigPackage = legacyinventory.SmithLintPackage
	LintConfigRange   = "^1.0.0"
)

// SmithLintConfig is a root ESLint config that is only smith's.
const SmithLintConfig = "import monosmith from '" + LintConfigPackage + "'\n\nexport default monosmith({ workspaceRoot: import.meta.dirname })\n"

// How a lint config became smith's.
const (
	// LintSwapped means the shared config's import and call were replaced and
	// every rule written beside them stays.
	LintSwapped = "swapped"
	// LintReplaced means the file is now smith's two lines: the config was
	// written by hand, and its rules are gone (the diff shows them).
	LintReplaced = "replaced"
)

var (
	sharedImport  = regexp.MustCompile(`(?m)^import\s+(\w+)\s+from\s+['"](@mnci/eslint-config)['"];?([ \t]*\r?)$`)
	callOptions   = regexp.MustCompile(`^\s*(\{\s*workspaceRoot\s*:\s*[\w.]+\s*,?\s*\})?\s*$`)
	alreadySmiths = regexp.MustCompile(`['"]` + regexp.QuoteMeta(LintConfigPackage) + `['"]`)
)

// SwapLintConfig makes a root ESLint config smith's. A config that imports a
// shared config smith replaces keeps the rules written beside it; any other
// is replaced whole. ok is false, with the reason, for a config whose call
// passes options smith's does not take.
func SwapLintConfig(text string) (rewritten, how, why string, ok bool) {
	if alreadySmiths.MatchString(text) {
		return text, "", "", true
	}
	m := sharedImport.FindStringSubmatch(text)
	if m == nil {
		return SmithLintConfig, LintReplaced, "", true
	}
	name := m[1]
	call := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\s*\(`)
	rewritten = sharedImport.ReplaceAllString(text, "import monosmith from '"+LintConfigPackage+"'${3}")
	var out strings.Builder
	rest := rewritten
	calls := 0
	for {
		loc := call.FindStringIndex(rest)
		if loc == nil {
			out.WriteString(rest)
			break
		}
		if inComment(rest, loc[0]) {
			out.WriteString(rest[:loc[0]] + "monosmith(")
			rest = rest[loc[1]:]
			continue
		}
		args, end, balanced := balancedArguments(rest[loc[1]:])
		if !balanced || !callOptions.MatchString(args) {
			return text, "", "the config calls " + m[2] + " with options smith's config does not take (" + strings.TrimSpace(args) + ")", false
		}
		out.WriteString(rest[:loc[0]] + "monosmith(")
		out.WriteString(args)
		out.WriteString(")")
		rest = rest[loc[1]+end+1:]
		calls++
	}
	if calls == 0 {
		return text, "", "the config imports " + m[2] + " but does not call it as a function", false
	}
	return out.String(), LintSwapped, "", true
}

// balancedArguments returns the text up to the parenthesis that closes a call
// whose opening one was just read, and the index of that parenthesis.
func balancedArguments(s string) (args string, end int, ok bool) {
	depth := 1
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[:i], i, true
			}
		}
	}
	return "", 0, false
}

// inComment reports whether the text at index starts on a line that is a
// comment (its first characters are //, /* or *).
func inComment(s string, index int) bool {
	start := strings.LastIndexByte(s[:index], 10) + 1
	line := strings.TrimSpace(s[start:index])
	return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "/*")
}
