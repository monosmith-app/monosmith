package toolinstall

import (
	"slices"
	"strings"
)

// Manager is a package manager smith can use to install a tool.
type Manager struct {
	// Name is the manager as plugins name it in ToolInstall.Manager.
	Name string `json:"name"`
	// Binary is the program that shows the manager is on the machine.
	Binary string `json:"binary"`
	// OS lists the systems (runtime.GOOS values) the manager is offered on.
	OS []string `json:"os"`
	// Language marks an installer that belongs to a language (uv, fnm, rustup,
	// npm): offered only when it is already on the machine.
	Language bool `json:"language,omitempty"`
}

// knownManagers are the managers, in the order they are offered.
var knownManagers = []Manager{
	{Name: "winget", Binary: "winget", OS: []string{"windows"}},
	{Name: "brew", Binary: "brew", OS: []string{"darwin", "linux"}},
	{Name: "brew-cask", Binary: "brew", OS: []string{"darwin"}},
	{Name: "apt", Binary: "apt-get", OS: []string{"linux"}},
	{Name: "dnf", Binary: "dnf", OS: []string{"linux"}},
	{Name: "pacman", Binary: "pacman", OS: []string{"linux"}},
	{Name: "uv", Binary: "uv", OS: []string{"windows", "darwin", "linux"}, Language: true},
	{Name: "fnm", Binary: "fnm", OS: []string{"windows", "darwin", "linux"}, Language: true},
	{Name: "rustup", Binary: "rustup", OS: []string{"windows", "darwin", "linux"}, Language: true},
	{Name: "npm", Binary: "npm", OS: []string{"windows", "darwin", "linux"}, Language: true},
}

// LookPath finds a program on PATH; exec.LookPath is the real one.
type LookPath func(program string) (string, error)

// DetectManagers returns the managers of a system that are on the machine.
//
// Parameters: goos is the system (runtime.GOOS); look finds a program on PATH.
// Returns: the managers in the order smith offers them.
func DetectManagers(goos string, look LookPath) []Manager {
	var out []Manager
	for _, m := range knownManagers {
		if !slices.Contains(m.OS, goos) {
			continue
		}
		if _, err := look(m.Binary); err == nil {
			out = append(out, m)
		}
	}
	return out
}

// Env is the system an install command is built for.
type Env struct {
	// OS is the system (runtime.GOOS).
	OS string
	// Root is true when smith already runs with administrator rights, so
	// sudo is not needed.
	Root bool
	// NonInteractive adds the flags that stop a manager from asking
	// questions; set only when nobody is at the terminal.
	NonInteractive bool
}

// InstallCommand is an install command ready to show and to run.
type InstallCommand struct {
	// Argv is the program and its arguments; no shell is involved.
	Argv []string
	// Admin is true when the command needs administrator rights, or may ask
	// the system for them (sudo, and winget, which can open the Windows
	// approval window).
	Admin bool
}

// String is the command as a person types it.
func (c InstallCommand) String() string { return strings.Join(c.Argv, " ") }

// BuildCommand builds the install command of a manager for a package. The
// package may name several packages separated by spaces ("nodejs npm").
// ok is false for a manager smith does not know.
func BuildCommand(manager, pkg string, env Env) (cmd InstallCommand, ok bool) {
	pkgs := strings.Fields(pkg)
	sudo := func(argv ...string) InstallCommand {
		if env.Root {
			return InstallCommand{Argv: argv}
		}
		return InstallCommand{Argv: append([]string{"sudo"}, argv...), Admin: true}
	}
	switch manager {
	case "winget":
		argv := []string{"winget", "install", "--id", pkg, "-e"}
		if env.NonInteractive {
			argv = append(argv, "--accept-package-agreements", "--accept-source-agreements", "--disable-interactivity")
		}
		return InstallCommand{Argv: argv, Admin: true}, true
	case "brew":
		return InstallCommand{Argv: append([]string{"brew", "install"}, pkgs...)}, true
	case "brew-cask":
		return InstallCommand{Argv: append([]string{"brew", "install", "--cask"}, pkgs...)}, true
	case "apt":
		argv := []string{"apt-get", "install"}
		if env.NonInteractive {
			argv = append(argv, "-y")
		}
		return sudo(append(argv, pkgs...)...), true
	case "dnf":
		argv := []string{"dnf", "install"}
		if env.NonInteractive {
			argv = append(argv, "-y")
		}
		return sudo(append(argv, pkgs...)...), true
	case "pacman":
		argv := []string{"pacman", "-S"}
		if env.NonInteractive {
			argv = append(argv, "--noconfirm")
		}
		return sudo(append(argv, pkgs...)...), true
	case "npm":
		return InstallCommand{Argv: append([]string{"npm", "install", "-g"}, pkgs...)}, true
	case "uv":
		return InstallCommand{Argv: []string{"uv", "python", "install", pkg}}, true
	case "fnm":
		return InstallCommand{Argv: []string{"fnm", "install", pkg}}, true
	case "rustup":
		return InstallCommand{Argv: []string{"rustup", "default", pkg}}, true
	}
	return InstallCommand{}, false
}
