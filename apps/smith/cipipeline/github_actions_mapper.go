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

permissions:
  contents: write
  pull-requests: write
%[2]s
concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}

jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
`, m.ReleaseBranch, idToken(m))
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	for _, a := range m.Authenticate {
		b.WriteString(githubAuthenticate(a))
	}
	fmt.Fprintf(&b, `      - name: Install smith
        run: curl -fsSL %s | sh -s -- "$SMITH_VERSION"
        env:
          SMITH_VERSION: %s
          GITHUB_TOKEN: ${{ github.token }}
`, InstallScriptURL, m.SmithVersion)
	b.WriteString(slot("      ", "before-setup"))
	b.WriteString("      - name: Setup\n        run: smith ci setup\n")
	b.WriteString(githubEnv("        ", m.SetupCredentials()))
	b.WriteString(`      - name: Audit
        run: smith ci audit
      - name: Verify
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
%[2]s      - uses: actions/upload-artifact@v7
        if: github.event_name == 'push' && github.ref_name == '%[1]s'
        with:
          name: drop
          path: dist/drop
          if-no-files-found: ignore
`, m.ReleaseBranch, githubEnv("        ", m.TaskCredentials()))
	}
	fmt.Fprintf(&b, `      - name: Release
        if: github.event_name == 'push' && github.ref_name == '%s'
        run: smith ci release
`, m.ReleaseBranch)
	if len(m.Credentials) > 0 {
		b.WriteString("        env:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "          %s: ${{ secrets.%s }}\n", c.Env, c.Env)
		}
	}
	if m.Native {
		b.WriteString(githubNativeJob(m))
	}
	return b.String()
}

// githubNativeJob builds the native projects on every OS after the ci job,
// so a release made there gets each platform's build attached.
func githubNativeJob(m Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, `
  native:
    needs: ci
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
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	for _, a := range m.Authenticate {
		b.WriteString(githubAuthenticate(a))
	}
	fmt.Fprintf(&b, `      - name: Install smith
        shell: bash
        run: curl -fsSL %s | sh -s -- "$SMITH_VERSION"
        env:
          SMITH_VERSION: %s
          GITHUB_TOKEN: ${{ github.token }}
      - name: Setup
        run: smith ci setup
`, InstallScriptURL, m.SmithVersion)
	b.WriteString(githubEnv("        ", m.SetupCredentials()))
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
	b.WriteString(`      - uses: actions/upload-artifact@v7
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
	if m.Provenance {
		b.WriteString("  id-token: write\n")
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
	fmt.Fprintf(&b, `      - uses: dart-lang/setup-dart@v1
      - name: Install smith
        run: curl -fsSL %s | sh -s -- "$SMITH_VERSION"
        env:
          SMITH_VERSION: %s
          GITHUB_TOKEN: ${{ github.token }}
`, InstallScriptURL, m.SmithVersion)
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
