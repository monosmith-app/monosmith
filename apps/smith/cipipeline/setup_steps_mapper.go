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
	fmt.Fprintf(&b, `      - name: Install smith
        run: curl -fsSL %s | sh -s -- "$SMITH_VERSION"
        env:
          SMITH_VERSION: %s
          GITHUB_TOKEN: ${{ github.token }}
      - name: Setup
        run: smith ci setup
`, InstallScriptURL, m.SmithVersion)
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
	fmt.Fprintf(&b, `  - bash: curl -fsSL %s | sh -s -- "%s"
    displayName: Install smith
  - bash: smith ci setup
    displayName: Setup
`, InstallScriptURL, m.SmithVersion)
	return b.String()
}
