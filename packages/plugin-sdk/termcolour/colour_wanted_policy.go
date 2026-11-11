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

// ToolColourEnv names the variable smith sets for the processes it starts
// (its plugins, and the child smith of the TUI): "1" when the tools they run
// stream to a person who sees colour, "0" when they must stay plain. Their own
// stdout is a pipe, so they cannot find this out alone.
const ToolColourEnv = "SMITH_TOOL_COLOR" // words:literal

// ToolWanted decides whether a tool smith runs may colour the output it
// streams to a person. NO_COLOR and TERM=dumb always win; then ToolColourEnv,
// when smith set it for this process; then Wanted.
//
// Parameters: getenv reads an environment variable; terminal reports whether
// stdout is a terminal.
// Returns: true when tools are asked for colour and smith paint may apply.
func ToolWanted(getenv func(string) string, terminal bool) bool {
	if getenv("NO_COLOR") != "" || getenv("TERM") == "dumb" {
		return false
	}
	switch getenv(ToolColourEnv) {
	case "1":
		return true
	case "0":
		return false
	}
	return Wanted(getenv, terminal)
}

// ToolEntry is the ToolColourEnv entry (KEY=VALUE) that hands the decision on
// to a child process.
//
// Parameters: on is the decision.
// Returns: "SMITH_TOOL_COLOR=1" or "SMITH_TOOL_COLOR=0".
func ToolEntry(on bool) string {
	if on {
		return ToolColourEnv + "=1"
	}
	return ToolColourEnv + "=0"
}
