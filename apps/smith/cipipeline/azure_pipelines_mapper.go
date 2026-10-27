package cipipeline

import (
	"fmt"
	"regexp"
	"strconv"
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

`, strings.Join(m.TriggerBranches(), ", "))
	b.WriteString(azureVariables(m))
	steps := azureSteps(m)
	if !m.Native {
		b.WriteString("pool:\n" + azurePool(m, "  ") + "\nsteps:\n")
		b.WriteString(steps)
		return b.String()
	}
	// A native project needs a second job, built on each OS: the steps move
	// under jobs.
	b.WriteString("jobs:\n  - job: ci\n    pool:\n" + azurePool(m, "      ") + "    steps:\n")
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
	b.WriteString(azureStores(m.Stores))
	b.WriteString(azureInstallSmith(m, "  "))
	b.WriteString(slot("  ", "before-setup"))
	b.WriteString("  - bash: smith ci setup\n    displayName: Setup\n")
	b.WriteString(azureEnv("    ", m.SetupCredentials()))
	b.WriteString(`  - bash: smith ci audit
    displayName: Audit
`)
	b.WriteString(azureEnv("    ", m.TaskCredentials()))
	b.WriteString(`  - bash: smith ci verify
    displayName: Verify
`)
	b.WriteString(azureEnv("    ", m.TaskCredentials()))
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
%[2]s  - publish: dist/drop
    artifact: drop
    displayName: Keep the packed apps (dist/drop as artifact drop)
    condition: %[1]s
`, onPush, azureEnv("    ", m.TaskCredentials()))
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
	b.WriteString(azureInstallSmith(m, "      "))
	b.WriteString("      - bash: smith ci setup\n        displayName: Setup\n")
	b.WriteString(azureEnv("        ", m.SetupCredentials()))
	b.WriteString(`      - bash: smith ci native
        displayName: Native
`)
	// The tasks it runs read the cache's credentials; a native package
	// publishes from here on a release push, which needs the release's too.
	b.WriteString(azureEnv("        ", m.nativeCredentials()))
	b.WriteString(`      - publish: dist/drop
        artifact: native-$(image)
        displayName: Keep the native builds (dist/drop as artifact native-OS)
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

// azurePool is the pool's setting, indented: the configured agent pool by
// name, else Microsoft's hosted ubuntu image.
func azurePool(m Model, indent string) string {
	if m.AzurePool != "" {
		return indent + "name: " + yamlWord(m.AzurePool) + "\n"
	}
	return indent + "vmImage: ubuntu-latest\n"
}

// azureVariables is the variables block: the linked variable groups, then the
// package stores' variables. Without groups it is the plain mapping.
func azureVariables(m Model) string {
	if len(m.AzureVariableGroups) == 0 {
		return azureStoreVariables(m.Stores)
	}
	var b strings.Builder
	b.WriteString("variables:\n")
	for _, g := range m.AzureVariableGroups {
		fmt.Fprintf(&b, "  - group: %s\n", yamlWord(g))
	}
	for _, s := range m.Stores {
		fmt.Fprintf(&b, "  - name: %s\n    value: %s/%s\n", s.Env, azureStoresDir, s.Name)
	}
	b.WriteString("\n")
	return b.String()
}

// plainYAMLWord is a name that needs no quotes as a YAML scalar.
var plainYAMLWord = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// yamlWord is a name as a YAML scalar: bare when it is plain, else quoted.
func yamlWord(s string) string {
	if plainYAMLWord.MatchString(s) {
		return s
	}
	return strconv.Quote(s)
}
