package updatenotice

import "strings"

// InstallCommand is the command that updates smith the way it was installed,
// read from where the running binary lives. goos is runtime.GOOS. A binary in
// an unknown place gets the release page.
func InstallCommand(exe, goos string) string {
	p := strings.ToLower(strings.ReplaceAll(exe, "\\", "/"))
	switch {
	case Managed(exe):
		return ManagedCommand
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

// ManagedCommand is what InstallCommand says for a smith a launcher started:
// smith update installs the new version itself, next to the old one.
const ManagedCommand = "smith update"

// Managed reports whether a smith binary lives in a versions folder the
// launchers and smith update fill (<smith home>/versions/<version>/), so
// smith updates by installing the next version there (ADR 0101).
//
// Parameters: exe is the running binary.
// Returns: true for a launcher-managed smith.
func Managed(exe string) bool {
	return strings.Contains(strings.ToLower(strings.ReplaceAll(exe, "\\", "/")), "/.smith/versions/")
}

// UpdateCommand is InstallCommand for a channel. The registry packages are
// launchers with their own version line, so an npm, pip or dotnet update
// installs the newest launcher on every channel, and the launcher runs the
// smith the channel names.
//
// Parameters: exe is the running binary, goos is runtime.GOOS, channel is
// "latest" or "next" (unused: every install command is the same on both).
// Returns: the command, or "see <releases page>" for an unknown install.
func UpdateCommand(exe, goos, _ string) string {
	return InstallCommand(exe, goos)
}

// isReleasePage is true for the answer of InstallCommand when it cannot tell
// how smith was installed.
func isReleasePage(command string) bool { return strings.HasPrefix(command, "see ") }
