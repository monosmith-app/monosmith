package dependencies

import (
	"regexp"
	"strings"
)

var constraintParts = regexp.MustCompile(`^(~>|>=|<=|>|<|=)?\s*(\d+(?:\.\d+)*)$`)

// Constraint writes the constraint a rewrite asks for the way RubyGems
// people write it: the operator, a space, the version. A pessimistic
// constraint keeps its precision: moving "~> 1.2" to 1.4.1 gives "~> 1.4",
// which still allows every 1.x from there, not "~> 1.4.1".
func Constraint(old, wanted string) string {
	w := constraintParts.FindStringSubmatch(strings.TrimSpace(wanted))
	if w == nil {
		return wanted
	}
	op, version := w[1], w[2]
	if o := constraintParts.FindStringSubmatch(strings.TrimSpace(old)); o != nil && op == "~>" {
		keep := len(strings.Split(o[2], "."))
		if parts := strings.Split(version, "."); len(parts) > keep {
			version = strings.Join(parts[:keep], ".")
		}
	}
	if op == "" {
		return version
	}
	return op + " " + version
}

// SetConstraint sets the version constraint of one gem in a gemspec or
// Gemfile source (the first quoted string after the gem's name) and says
// whether it found the gem. Only that string changes: its quotes, the rest
// of the line and the file's formatting stay as written.
func SetConstraint(src, name, constraint string) (string, bool) {
	line := regexp.MustCompile(`(?m)^(\s*(?:\w+\.add_(?:runtime_|development_)?dependency\s*\(?\s*|gem\s+\(?\s*)(['"])` +
		regexp.QuoteMeta(name) + `['"]\s*,\s*(['"]))[^'"]*(['"])`)
	if !line.MatchString(src) {
		return src, false
	}
	done := false
	out := line.ReplaceAllStringFunc(src, func(match string) string {
		if done {
			return match
		}
		done = true
		parts := line.FindStringSubmatch(match)
		return parts[1] + constraint + parts[4]
	})
	return out, true
}
