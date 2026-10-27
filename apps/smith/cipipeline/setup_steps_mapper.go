package cipipeline

import (
	"fmt"
	"strings"
)

// GitHubSetup is how every job smith writes starts, as GitHub Actions steps
// (at a job's steps indentation): checkout with full history, the
// toolchains, smith at the version that wrote it, and smith ci setup.
func GitHubSetup(m Model) string {
	var b strings.Builder
	b.WriteString(`      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
`)
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	b.WriteString(githubInstallSmith(m, false))
	b.WriteString("      - name: Setup\n        run: smith ci setup\n")
	return b.String()
}

// AzureSetup is the same start as Azure Pipelines steps (at the top level's
// indentation).
func AzureSetup(m Model) string {
	var b strings.Builder
	b.WriteString(`  - checkout: self
    fetchDepth: 0
`)
	for _, t := range m.Toolchains {
		b.WriteString(azureSetup(t))
	}
	b.WriteString(azureInstallSmith(m, "  "))
	b.WriteString("  - bash: smith ci setup\n    displayName: Setup\n")
	return b.String()
}

// githubInstallSmith is the GitHub Actions step that installs smith (see
// SmithInstallFor). bash makes the script step portable to the Windows
// runners of the native job; the registry step is always bash.
func githubInstallSmith(m Model, bash bool) string {
	i := SmithInstallFor(m)
	if !i.Script {
		command := i.HostedCommand()
		if dir := i.HostedPathDir(); dir != "" {
			command += "\necho \"" + dir + "\" >> \"$GITHUB_PATH\""
		}
		return "      - name: Install smith\n        shell: bash\n        run: " + runValue("        ", command)
	}
	shell := ""
	if bash {
		shell = "        shell: bash\n"
	}
	return fmt.Sprintf(`      - name: Install smith
%s        run: sh %s "$SMITH_VERSION"
        env:
          SMITH_VERSION: %s
          GITHUB_TOKEN: ${{ github.token }}
`, shell, InstallScriptPath, m.SmithVersion)
}

// azureInstallSmith is the Azure Pipelines step that installs smith, at the
// given indentation of its "- bash:" line.
func azureInstallSmith(m Model, indent string) string {
	i := SmithInstallFor(m)
	if !i.Script {
		command := i.HostedCommand()
		if dir := i.HostedPathDir(); dir != "" {
			command += "\necho \"##vso[task.prependpath]" + dir + "\""
		}
		return indent + "- bash: " + runValue(indent+"  ", command) + indent + "  displayName: Install smith\n"
	}
	return fmt.Sprintf("%s- bash: sh %s \"%s\"\n%s  displayName: Install smith\n", indent, InstallScriptPath, m.SmithVersion, indent)
}
