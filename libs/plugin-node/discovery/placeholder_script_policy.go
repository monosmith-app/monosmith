package discovery

import "regexp"

// placeholderScript matches the test script npm init writes,
// echo "Error: no test specified" && exit 1, and its close variants (other
// quotes, ; or || before exit, another exit code, no exit at all).
var placeholderScript = regexp.MustCompile(`(?i)^\s*echo\s+\\?["']?\s*error:\s*no test specified\s*\\?["']?\s*((&&|\|\||;)\s*exit\s+\d+\s*)?$`)

// IsPlaceholderScript reports whether a package.json script is a placeholder
// that npm init writes, not a real command. Such a script gives no task: it
// only fails, so smith treats the project as one without tests.
//
// Parameters: command is the script's command line.
// Returns: true for the npm-init test placeholder and its close variants.
func IsPlaceholderScript(command string) bool {
	return placeholderScript.MatchString(command)
}
