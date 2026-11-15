package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// githubSetup renders the steps that install a toolchain on GitHub Actions:
// the plugin's action, else its install command, else a comment saying the
// plugin gave neither.
func githubSetup(t protocol.Toolchain) string {
	switch {
	case t.GitHub != nil:
		return actionStep("      ", "- uses: ", "with", t.GitHub)
	case t.Install != "":
		// bash on every runner: an install script is shell, and the Windows
		// runners' default is PowerShell.
		return fmt.Sprintf("      - name: Install %s\n        shell: bash\n        run: %s", t.Name, runValue("        ", t.Install))
	}
	return missing("      ", t)
}

// azureSetup renders the steps that install a toolchain on Azure Pipelines.
func azureSetup(t protocol.Toolchain) string {
	switch {
	case t.Azure != nil:
		return actionStep("  ", "- task: ", "inputs", t.Azure)
	case t.Install != "":
		return fmt.Sprintf("  - bash: %s    displayName: Install %s\n", runValue("    ", t.Install), t.Name)
	}
	return missing("  ", t)
}

// githubAuthenticate renders a registry sign-in step on GitHub Actions, or
// nothing: there the credentials come from the environment.
func githubAuthenticate(a protocol.Toolchain) string {
	if a.GitHub == nil {
		return ""
	}
	return actionStep("      ", "- uses: ", "with", a.GitHub)
}

// azureAuthenticate renders a registry sign-in task on Azure Pipelines
// (npmAuthenticate@0, PipAuthenticate@1, NuGetAuthenticate@1), which signs
// in with the pipeline's own token.
func azureAuthenticate(a protocol.Toolchain) string {
	if a.Azure == nil {
		return ""
	}
	return actionStep("  ", "- task: ", "inputs", a.Azure)
}

// githubEnv renders an env block of secrets, or nothing.
func githubEnv(indent string, creds []protocol.Credential) string {
	if len(creds) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(indent + "env:\n")
	for _, c := range creds {
		fmt.Fprintf(&b, "%s  %s: ${{ secrets.%s }}\n", indent, c.Env, c.Env)
	}
	return b.String()
}

// azureEnv renders an env block of pipeline variables, or nothing.
func azureEnv(indent string, creds []protocol.Credential) string {
	if len(creds) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(indent + "env:\n")
	for _, c := range creds {
		fmt.Fprintf(&b, "%s  %s: $(%s)\n", indent, c.Env, c.Env)
	}
	return b.String()
}

func actionStep(indent, head, inputsKey string, s *protocol.ActionStep) string {
	var b strings.Builder
	b.WriteString(indent + head + s.Uses + "\n")
	if len(s.Inputs) > 0 {
		b.WriteString(indent + "  " + inputsKey + ":\n")
		for _, in := range s.Inputs {
			b.WriteString(indent + "    " + in.Name + ": " + yamlValue(in.Value) + "\n")
		}
	}
	return b.String()
}

// runValue is a run/bash value: inline for one line, a block for several.
func runValue(indent, command string) string {
	command = strings.TrimRight(command, "\n")
	if !strings.Contains(command, "\n") {
		return command + "\n"
	}
	return "|\n" + indent + "  " + strings.ReplaceAll(command, "\n", "\n"+indent+"  ") + "\n"
}

func yamlValue(v string) string {
	if v == "true" || v == "false" {
		return v
	}
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

func missing(indent string, t protocol.Toolchain) string {
	return fmt.Sprintf("%s# smith: no plugin says how to install %s %s here; add a step in the before-setup slot\n", indent, t.Name, t.Version)
}
