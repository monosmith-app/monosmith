package cipipeline

import "strings"

// SmithInstall is how a generated pipeline gets smith onto the agent. This is
// the one place that rule lives.
//
// monosmith is published at the same version on npm, PyPI and NuGet, so a
// workspace whose toolchains include one of those runtimes installs smith with
// that runtime's package manager and writes no script into the repository:
// Node first (npm), then Python (pip), then .NET (dotnet tool). Every other
// workspace (Go, Rust, JVM, ...) keeps the committed InstallScriptPath, which
// downloads the release binary and verifies its checksum.
//
// The registry install is pinned to the same version as the script ([ci]
// smith_version, else the running smith); "latest" (a dev build) is unpinned.
// npm and pip name the public registry, so a private feed that
// npmAuthenticate or PipAuthenticate points the agent at cannot hide it.
type SmithInstall struct {
	// Script is true when the committed install script is used.
	Script bool
	tool   string
	pin    string
}

// SmithInstallFor picks the install for a pipeline model.
func SmithInstallFor(m Model) SmithInstall {
	pin := m.SmithVersion
	if pin == "latest" {
		pin = ""
	}
	for _, tool := range []string{"node", "python", "dotnet"} {
		if hasToolchain(m.Toolchains, tool) {
			return SmithInstall{tool: tool, pin: pin}
		}
	}
	return SmithInstall{Script: true}
}

// HostedCommand installs smith on a GitHub or Azure agent, where the
// toolchain's own step already put its bin directory on PATH (so pip needs no
// --user there).
func (i SmithInstall) HostedCommand() string { return i.command(false) }

// ContainerCommand installs smith in a container the pipeline provisions
// itself (GitLab, CircleCI, dev container), where pip installs per user.
func (i SmithInstall) ContainerCommand() string { return i.command(true) }

// HostedPathDir is the directory to add to PATH after HostedCommand, as a
// bash expression, or "" when the install already lands on PATH.
func (i SmithInstall) HostedPathDir() string {
	if i.tool == "dotnet" {
		// USERPROFILE on Windows runners (bash there), HOME elsewhere.
		return `${USERPROFILE:-$HOME}/.dotnet/tools`
	}
	return ""
}

// ContainerPathDir is the directory to add to PATH after ContainerCommand, or
// "" when the install already lands on PATH. Every container pipeline exports
// $HOME/.local/bin already, which is where pip --user installs.
func (i SmithInstall) ContainerPathDir() string {
	if i.tool == "dotnet" {
		return "$HOME/.dotnet/tools"
	}
	return ""
}

func (i SmithInstall) command(container bool) string {
	var b strings.Builder
	switch i.tool {
	case "node":
		b.WriteString("npm install --global --registry https://registry.npmjs.org/ monosmith")
		if i.pin != "" {
			b.WriteString("@" + shellWord(i.pin))
		}
	case "python":
		// Windows runners have python, not python3; the containers python3.
		python := "python"
		if container {
			python = "python3"
		}
		b.WriteString(python + " -m pip install --disable-pip-version-check --index-url https://pypi.org/simple/ ")
		if container {
			b.WriteString("--user ")
		}
		b.WriteString("monosmith")
		if i.pin != "" {
			b.WriteString("==" + shellWord(i.pin))
		}
	case "dotnet":
		b.WriteString("dotnet tool update --global monosmith")
		if i.pin != "" {
			b.WriteString(" --version " + shellWord(i.pin))
		}
	}
	return b.String()
}
