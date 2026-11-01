package cipipeline

import (
	"fmt"
	"strings"
)

// GitHubPath is where the GitHub Actions workflow is written.
const GitHubPath = ".github/workflows/ci.yml"

// RenderGitHub renders the model as a GitHub Actions workflow.
func RenderGitHub(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	fmt.Fprintf(&b, `name: ci

on:
  push:
    branches: [%[1]s]
  pull_request:
    branches: [%[1]s]
%[4]s
permissions:
  contents: write
  pull-requests: write
%[2]s
concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}

jobs:
  ci:
%[5]s    runs-on: ubuntu-latest
%[3]s    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
`, strings.Join(m.TriggerBranches(), ", "), idToken(m), githubCIOutputs(m), githubAssetsDispatch(m), githubCIOnlyWithoutAssetsTag(m))
	b.WriteString(githubSiblingCheckouts(m))
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	for _, a := range m.Authenticate {
		b.WriteString(githubAuthenticate(a))
	}
	b.WriteString(githubStores(m.Stores))
	b.WriteString(githubInstallSmith(m, false))
	b.WriteString(githubSiblingPreBuilds(m))
	b.WriteString(slot("      ", "before-setup"))
	b.WriteString("      - name: Setup\n        run: smith ci setup\n")
	b.WriteString(githubEnv("        ", m.SetupCredentials()))
	b.WriteString(`      - name: Audit
        run: smith ci audit
`)
	b.WriteString(githubEnv("        ", m.TaskCredentials()))
	b.WriteString(`      - name: Verify
        run: smith ci verify
`)
	b.WriteString(githubEnv("        ", m.TaskCredentials()))
	b.WriteString(slot("      ", "after-verify"))
	b.WriteString(`      - name: Release preview
        if: github.event_name == 'pull_request'
        run: smith release --dry-run --comment
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`)
	if m.Pack {
		fmt.Fprintf(&b, `      - name: Pack
        if: github.event_name == 'push' && github.ref_name == '%[1]s'
        run: smith ci pack
%[2]s      - name: Keep the packed apps (dist/drop as artifact drop)
        uses: actions/upload-artifact@v7
        if: github.event_name == 'push' && github.ref_name == '%[1]s'
        with:
          name: drop
          path: dist/drop
          if-no-files-found: ignore
`, m.ReleaseBranch, githubEnv("        ", m.TaskCredentials()))
	}
	b.WriteString(githubRubyGemsSignIn(m, "ci"))
	b.WriteString(githubAzureSignIn(m, "ci"))
	releaseID := ""
	if githubHasDependents(m) {
		releaseID = "        id: release\n"
	}
	fmt.Fprintf(&b, `      - name: Release
%[2]s        if: github.event_name == 'push' && %[1]s
        run: smith ci release
`, githubOnReleaseBranches(m), releaseID)
	if len(m.Credentials) > 0 || m.VSCEEntra {
		b.WriteString("        env:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "          %s: ${{ secrets.%s }}\n", c.Env, c.Env)
		}
		b.WriteString(githubEntraEnv(m))
	}
	// No marker step follows the release: smith ci release writes released=true to
	// the step's output itself (markReleased), the moment a project is released, so
	// a later step that fails does not keep the native and assets jobs from running
	// on a release that exists, and a push that releases nothing starts neither.
	if m.Attest {
		continueOnError := ""
		if m.AttestOptional {
			continueOnError = "        continue-on-error: true\n"
		}
		fmt.Fprintf(&b, `      - name: Attest build provenance
        if: github.event_name == 'push' && github.ref_name == '%s' && hashFiles('dist/release/**') != ''
%s        uses: actions/attest-build-provenance@v3
        with:
          subject-path: dist/release/*/*
`, m.ReleaseBranch, continueOnError)
	}
	if m.Native {
		b.WriteString(githubNativeJob(m))
	}
	b.WriteString(githubAssetsJob(m))
	return b.String()
}

// githubRubyGemsSignIn signs in to RubyGems with the run's OIDC token (RubyGems
// trusted publishing) right before the step that publishes, and only on a
// push to the release branch: the key it writes is short-lived and must not
// be minted for a pull request. It exports GEM_HOST_API_KEY for gem push.
func githubRubyGemsSignIn(m Model, job string) string {
	if !m.RubyGemsTrusted || (job == "native" && !m.NativePublish) {
		return ""
	}
	return fmt.Sprintf(`      - name: Sign in to RubyGems
        if: github.event_name == 'push' && %s
        uses: rubygems/configure-rubygems-credentials@v2.1.0
`, githubOnReleaseBranches(m))
}

// githubAzureSignIn signs in to Azure with the run's OIDC token (workload
// identity federation) right before the step that publishes, and only on a
// push to the release branch, so vsce can publish to the Marketplace as the
// app registration instead of with a PAT. The client and tenant ids are the
// repository secrets AZURE_CLIENT_ID and AZURE_TENANT_ID; no subscription is
// needed.
func githubAzureSignIn(m Model, job string) string {
	if !m.VSCEEntra || (job == "native" && !m.NativePublish) {
		return ""
	}
	return fmt.Sprintf(`      - name: Sign in to Azure
        if: github.event_name == 'push' && %s
        uses: azure/login@v3
        with:
          client-id: ${{ secrets.AZURE_CLIENT_ID }}
          tenant-id: ${{ secrets.AZURE_TENANT_ID }}
          allow-no-subscriptions: true
`, githubOnReleaseBranches(m))
}

// githubEntraEnv tells the release to publish extensions with the Azure
// sign-in: vsce then uses its Entra ID credential rather than VSCE_PAT.
func githubEntraEnv(m Model) string {
	if !m.VSCEEntra {
		return ""
	}
	return "          VSCE_AZURE_CREDENTIAL: \"true\"\n"
}

// githubNativeJob builds the native projects on every OS after the ci job,
// so a release made there gets each platform's build attached.
func githubNativeJob(m Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, `
  native:
    needs: ci
    if: ${{ always() && (needs.ci.result == 'success' || (github.event_name == 'push' && needs.ci.outputs.released == 'true')) }}
    strategy:
      fail-fast: false
      matrix:
        os: [%s]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
`, strings.Join(m.NativeRunners, ", "))
	b.WriteString(githubSiblingCheckouts(m))
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	for _, a := range m.Authenticate {
		b.WriteString(githubAuthenticate(a))
	}
	b.WriteString(githubInstallSmith(m, true))
	b.WriteString(githubSiblingPreBuilds(m))
	b.WriteString("      - name: Setup\n        run: smith ci setup\n")
	b.WriteString(githubEnv("        ", m.SetupCredentials()))
	b.WriteString(githubRubyGemsSignIn(m, "native"))
	b.WriteString(githubAzureSignIn(m, "native"))
	b.WriteString(`      - name: Native
        run: smith ci native
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`)
	// The tasks it runs read the cache's credentials; a native package
	// publishes from here on a release push, which needs the release's too.
	for _, c := range m.nativeCredentials() {
		fmt.Fprintf(&b, "          %s: ${{ secrets.%s }}\n", c.Env, c.Env)
	}
	if m.NativePublish {
		b.WriteString(githubEntraEnv(m))
	}
	b.WriteString(`      - name: Keep the native builds (dist/drop as artifact native-OS)
        uses: actions/upload-artifact@v7
        with:
          name: native-${{ matrix.os }}
          path: dist/drop
          if-no-files-found: ignore
`)
	return b.String()
}

// idToken lets the job request an OIDC token, which npm provenance signs
// the published packages with; actions: write lets the release start the
// pub workflow.
func idToken(m Model) string {
	var b strings.Builder
	if m.Pub {
		b.WriteString("  actions: write\n")
	}
	if m.Provenance || m.Attest || m.RubyGemsTrusted || m.VSCEEntra {
		b.WriteString("  id-token: write\n")
	}
	if m.Attest {
		b.WriteString("  attestations: write\n")
	}
	// Plugins add what their publishing needs: packages: write to push
	// images to ghcr.io.
	for _, perm := range m.Permissions {
		b.WriteString("  " + perm + "\n")
	}
	return b.String()
}

// GitHubPubPath is the workflow that publishes to pub.dev.
const GitHubPubPath = ".github/workflows/pub.yml"

// RenderGitHubPub renders the pub workflow. The release starts it on an
// app's pub tag (pub.dev's automated publishing accepts only a run on a tag
// matching the package's pattern); setup-dart gives dart pub.dev's OIDC
// token, and smith builds the package from the GitHub Release's archives.
func RenderGitHubPub(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	b.WriteString(`name: pub

on:
  workflow_dispatch:
    inputs:
      project:
        description: The released app to publish
        required: true
      version:
        description: Its released version
        required: true

permissions:
  contents: read
  id-token: write

jobs:
  pub:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
`)
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	b.WriteString("      - uses: dart-lang/setup-dart@v1\n")
	b.WriteString(githubInstallSmith(m, false))
	b.WriteString(slot("      ", "before-publish"))
	b.WriteString(`      - name: Publish to pub.dev
        run: smith release pub --project "$PROJECT" --version "$VERSION"
        env:
          PROJECT: ${{ inputs.project }}
          VERSION: ${{ inputs.version }}
          GITHUB_TOKEN: ${{ github.token }}
`)
	return b.String()
}

func header(comment string) string {
	return comment + " Generated by smith from the workspace's projects. `smith add` and `smith ci generate`\n" +
		comment + " rewrite this file: put your own steps between the smith:slot markers, which are kept.\n"
}

func slot(indent, name string) string {
	return indent + "# smith:slot " + name + "\n" + indent + "# smith:slot-end " + name + "\n"
}

// githubHasDependents says whether jobs run after the ci job: the native jobs
// and the assets jobs.
func githubHasDependents(m Model) bool { return m.Native || len(m.AssetPlatforms) > 0 }

// githubCIOutputs is the ci job's outputs block: whether the release step went
// through, for the jobs that need the release to exist rather than the whole
// job to succeed. Empty when no job depends on the ci job.
func githubCIOutputs(m Model) string {
	if !githubHasDependents(m) {
		return ""
	}
	return "    outputs:\n      released: ${{ steps.release.outputs.released }}\n"
}

// githubOnReleaseBranches is the condition that holds on a push to the release
// branch or to a release channel's branch: where the release runs, and so where
// the sign-in steps that mint its short-lived credentials run.
func githubOnReleaseBranches(m Model) string {
	return m.branchCondition(func(branch string) string { return "github.ref_name == '" + branch + "'" }, func(c []string) string { return "(" + strings.Join(c, " || ") + ")" })
}
