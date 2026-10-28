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
