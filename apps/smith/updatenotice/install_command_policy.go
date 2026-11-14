package updatenotice

import "strings"

// InstallCommand is the command that updates smith the way it was installed,
// read from where the running binary lives. goos is runtime.GOOS. A binary in
// an unknown place gets the release page.
func InstallCommand(exe, goos string) string {
	p := strings.ToLower(strings.ReplaceAll(exe, "\\", "/"))
	switch {
	case strings.Contains(p, "/node_modules/"):
		return "npm i -g monosmith@latest"
	case strings.Contains(p, "/.dotnet/tools"):
		return "dotnet tool update --global monosmith"
	case strings.Contains(p, "/uv/tools/"):
		return "uv tool upgrade monosmith"
	case strings.Contains(p, "/pipx/"):
		return "pipx upgrade monosmith"
	case strings.Contains(p, "/site-packages/") || (strings.Contains(p, "/python") && strings.Contains(p, "/scripts/")):
		return "pip install --upgrade monosmith"
	case strings.Contains(p, "/.local/bin/"):
		if goos == "windows" {
			return "irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1 | iex"
		}
		return "curl -fsSL https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.sh | sh"
	}
	return "see " + ReleasesURL
}

// UpdateCommand is InstallCommand for a channel: on next, an npm command
// installs the next tag instead of latest.
//
// Parameters: exe is the running binary, goos is runtime.GOOS, channel is
// "latest" or "next".
// Returns: the command, or "see <releases page>" for an unknown install.
func UpdateCommand(exe, goos, channel string) string {
	command := InstallCommand(exe, goos)
	if channel == "next" {
		command = strings.Replace(command, "@latest", "@next", 1)
	}
	return command
}

// isReleasePage is true for the answer of InstallCommand when it cannot tell
// how smith was installed.
func isReleasePage(command string) bool { return strings.HasPrefix(command, "see ") }
