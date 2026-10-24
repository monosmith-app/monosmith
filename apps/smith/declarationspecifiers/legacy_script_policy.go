package declarationspecifiers

import "regexp"

// LegacyFile is the workspace-root script smith convert used to write, and
// that builds converted by it still run as `node ../../declaration-specifiers.mjs`.
const LegacyFile = "declaration-specifiers.mjs"

var legacyCall = regexp.MustCompile(`node\s+(?:\.\./)*` + regexp.QuoteMeta(LegacyFile))

// MigrateScript replaces a call of the legacy root script in a package.json
// script with the smith command, and says whether it found one.
func MigrateScript(script string) (string, bool) {
	if !legacyCall.MatchString(script) {
		return script, false
	}
	return legacyCall.ReplaceAllString(script, "smith "+Name), true
}
