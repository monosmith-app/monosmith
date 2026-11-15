package shellenv

import (
	"path"
	"strings"
)

// Shell is a shell that smith knows how to complete in.
type Shell string

// The supported shells.
const (
	Bash       Shell = "bash"
	Zsh        Shell = "zsh"
	Fish       Shell = "fish"
	PowerShell Shell = "powershell"
)

// Shells lists every supported shell, in the order the status shows them.
var Shells = []Shell{Bash, Zsh, Fish, PowerShell}

// ParseShell reads the name of a shell or of its program.
//
// Parameters: name is a shell name, a program name or a path to a program
// ("zsh", "-bash", "/usr/bin/fish", "pwsh.exe").
// Returns: the shell, and false when the name is not a supported shell.
func ParseShell(name string) (Shell, bool) {
	base := path.Base(strings.ReplaceAll(strings.TrimSpace(name), `\`, "/"))
	base = strings.TrimSuffix(strings.ToLower(strings.TrimPrefix(base, "-")), ".exe")
	switch base {
	case "bash":
		return Bash, true
	case "zsh":
		return Zsh, true
	case "fish":
		return Fish, true
	case "pwsh", "powershell":
		return PowerShell, true
	}
	return "", false
}

// Title is the name of the shell as a person writes it.
func (s Shell) Title() string {
	switch s {
	case PowerShell:
		return "PowerShell"
	case Bash:
		return "Bash"
	case Zsh:
		return "Zsh"
	case Fish:
		return "Fish"
	}
	return string(s)
}
