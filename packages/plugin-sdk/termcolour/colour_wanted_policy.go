package termcolour

import "strings"

// Wanted decides, in this one place, whether output meant for a person
// carries ANSI colour. GitHub Actions, Azure Pipelines and GitLab render it;
// elsewhere it is for a terminal only. NO_COLOR and TERM=dumb always win;
// FORCE_COLOR turns it on for a captured sample.
//
// Parameters: getenv reads an environment variable (os.Getenv in production);
// terminal reports whether stdout is a terminal.
// Returns: true when the output may carry colour codes.
func Wanted(getenv func(string) string, terminal bool) bool {
	switch {
	case getenv("NO_COLOR") != "" || getenv("TERM") == "dumb":
		return false
	case getenv("FORCE_COLOR") != "" && getenv("FORCE_COLOR") != "0":
		return true
	case getenv("GITHUB_ACTIONS") == "true", strings.EqualFold(getenv("TF_BUILD"), "true"), getenv("GITLAB_CI") != "":
		return true
	}
	return terminal
}
