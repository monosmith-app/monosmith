package doctor

import "strings"

// IsMachineCheck reports whether a failed check is about the computer smith
// runs on (a tool that is missing, or older than the one CI installs) and not
// about a file a command wrote. No file can cause such a check, so a command
// that wrote files never answers for it.
func IsMachineCheck(f Finding) bool {
	check := strings.ToLower(f.Check)
	detail := strings.ToLower(f.Detail)
	return strings.HasSuffix(check, " is installed") ||
		strings.Contains(check, ", as in ci") ||
		strings.Contains(detail, "is not on path") ||
		strings.Contains(detail, "is not installed")
}
