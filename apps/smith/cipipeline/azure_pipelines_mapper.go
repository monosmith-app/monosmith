package cipipeline

import (
	"fmt"
	"strings"
)

// AzurePath is where the Azure Pipelines definition is written.
const AzurePath = "azure-pipelines.yml"

// RenderAzure renders the model as an Azure Pipelines definition.
func RenderAzure(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	fmt.Fprintf(&b, `trigger:
  branches:
    include: [%[1]s]

pr:
  branches:
    include: [%[1]s]

`, m.ReleaseBranch)
	steps := azureSteps(m)
	if !m.Native {
		b.WriteString("pool:\n  vmImage: ubuntu-latest\n\nsteps:\n")
		b.WriteString(steps)
		return b.String()
	}
	// A native project needs a second job, built on each OS: the steps move
	// under jobs.
	b.WriteString("jobs:\n  - job: ci\n    pool:\n      vmImage: ubuntu-latest\n    steps:\n")
	b.WriteString(indent(steps, "    "))
	b.WriteString(azureNativeJob(m))
	return b.String()
}

// azureSteps are the ci job's steps, at the top level's indentation.
func azureSteps(m Model) string {
	var b strings.Builder
	b.WriteString(`  - checkout: self
    fetchDepth: 0
    persistCredentials: true
`)
	for _, t := range m.Toolchains {
		b.WriteString(azureSetup(t))
	}
	for _, a := range m.Authenticate {
		b.WriteString(azureAuthenticate(a))
	}
	fmt.Fprintf(&b, `  - bash: curl -fsSL %s | sh -s -- "%s"
    displayName: Install smith
`, InstallScriptURL, m.SmithVersion)
	b.WriteString(slot("  ", "before-setup"))
	b.WriteString("  - bash: smith ci setup\n    displayName: Setup\n")
	b.WriteString(azureEnv("    ", m.SetupCredentials()))
	b.WriteString(`  - bash: smith ci audit
    displayName: Audit
  - bash: smith ci verify
    displayName: Verify
`)
	b.WriteString(slot("  ", "after-verify"))
	b.WriteString(`  - bash: smith release --dry-run --comment
    displayName: Release preview
    condition: and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'))
    env:
      SYSTEM_ACCESSTOKEN: $(System.AccessToken)
`)
	onPush := fmt.Sprintf("and(succeeded(), in(variables['Build.Reason'], 'IndividualCI', 'BatchedCI'), eq(variables['Build.SourceBranch'], 'refs/heads/%s'))", m.ReleaseBranch)
	if m.Pack {
		fmt.Fprintf(&b, `  - bash: smith ci pack
    displayName: Pack
    condition: %[1]s
  - publish: dist/drop
    artifact: drop
    displayName: Keep the packed apps
    condition: %[1]s
`, onPush)
	}
	fmt.Fprintf(&b, `  - bash: smith ci release
    displayName: Release
    condition: %s
`, onPush)
	if len(m.Credentials) > 0 {
		b.WriteString("    env:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "      %s: $(%s)\n", c.Env, c.Env)
		}
	}
	return b.String()
}

// azureNativeJob builds the native projects on every runner image after the
// ci job.
func azureNativeJob(m Model) string {
	var b strings.Builder
	b.WriteString("  - job: native\n    dependsOn: ci\n    strategy:\n      matrix:\n")
	for _, r := range m.NativeRunners {
		fmt.Fprintf(&b, "        %s:\n          image: %s\n", strings.ReplaceAll(r, "-", "_"), r)
	}
	b.WriteString("    pool:\n      vmImage: $(image)\n    steps:\n      - checkout: self\n        fetchDepth: 0\n")
	var tools strings.Builder
	for _, t := range m.Toolchains {
		tools.WriteString(azureSetup(t))
	}
	for _, a := range m.Authenticate {
		tools.WriteString(azureAuthenticate(a))
	}
	b.WriteString(indent(tools.String(), "    "))
	fmt.Fprintf(&b, `      - bash: curl -fsSL %s | sh -s -- "%s"
        displayName: Install smith
      - bash: smith ci setup
        displayName: Setup
`, InstallScriptURL, m.SmithVersion)
	b.WriteString(azureEnv("        ", m.SetupCredentials()))
	b.WriteString(`      - bash: smith ci native
        displayName: Native
`)
	// A native package publishes from here on a release push: the job needs
	// the release credentials too.
	if m.NativePublish && len(m.Credentials) > 0 {
		b.WriteString("        env:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "          %s: $(%s)\n", c.Env, c.Env)
		}
	}
	b.WriteString(`      - publish: dist/drop
        artifact: native-$(image)
        displayName: Keep the native builds
`)
	return b.String()
}

// indent prefixes every non-empty line.
func indent(text, prefix string) string {
	lines := strings.SplitAfter(text, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "")
}
