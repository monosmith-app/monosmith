package termcolour

// colourEnv asks each tool smith's plugins run to colour its output although
// its stdout is a pipe. Each entry names the tools known to honour it.
var colourEnv = []string{
	// An inherited NO_COLOR (from an outer plain smith) is cleared: the
	// convention reads an empty value as unset.
	"NO_COLOR=",
	// Node: chalk and supports-color (jest, vitest, eslint, pnpm, yarn, mocha);
	// Symfony Console 6.2+ (composer); many others.
	"FORCE_COLOR=1",
	// The CLICOLOR convention: Rust's anstream (ruff, uv), CMake, Ninja.
	"CLICOLOR_FORCE=1",
	// npm's own output (its color setting defaults to "only on a terminal").
	"npm_config_color=always",
	// cargo, and rustc and clippy through it.
	"CARGO_TERM_COLOR=always",
	// pytest; PYTHON_COLORS for the tracebacks of Python 3.13+.
	"PY_COLORS=1",
	"PYTHON_COLORS=1",
	"MYPY_FORCE_COLOR=1",
	// .NET writes Console colours as ANSI when its output is redirected
	// (dotnet build, dotnet test).
	"DOTNET_SYSTEM_CONSOLE_ALLOW_ANSI_COLOR_REDIRECTION=1",
}

// plainEnv turns each of those off again, and adds NO_COLOR, for a tool
// whose output smith parses. An inherited FORCE_COLOR would otherwise win.
var plainEnv = []string{
	"NO_COLOR=1",
	"FORCE_COLOR=0",
	"CLICOLOR_FORCE=0",
	"npm_config_color=false",
	"CARGO_TERM_COLOR=never",
	"PY_COLORS=0",
	"PYTHON_COLORS=0",
	"MYPY_FORCE_COLOR=0",
	"DOTNET_SYSTEM_CONSOLE_ALLOW_ANSI_COLOR_REDIRECTION=0",
}

// Env returns the environment entries (KEY=VALUE) that ask a tool for
// colour, or for plain text. Append them after the inherited environment:
// os/exec keeps the last value of a repeated key.
//
// Parameters: on chooses colour (true) or plain text (false).
// Returns: a new slice the caller may append to.
func Env(on bool) []string {
	if on {
		return append([]string(nil), colourEnv...)
	}
	return append([]string(nil), plainEnv...)
}
