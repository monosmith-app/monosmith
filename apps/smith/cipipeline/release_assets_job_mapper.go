package cipipeline

import (
	"fmt"
	"strings"
)

// githubAssetRunners is the GitHub-hosted runner of each platform a release
// build command runs for.
var githubAssetRunners = map[string]string{
	"windows-x64": "windows-latest",
	"macos-arm64": "macos-latest",
	"macos-x64":   "macos-15-intel",
	"linux-x64":   "ubuntu-latest",
	"linux-arm64": "ubuntu-24.04-arm",
}

// githubAssetsDispatch is the workflow_dispatch trigger that runs the assets
// job again for a release that exists: its assets_tag input is the tag
// (project@version). Empty when the workflow has no assets job.
func githubAssetsDispatch(m Model) string {
	if len(m.AssetPlatforms) == 0 {
		return ""
	}
	return `  workflow_dispatch:
    inputs:
      assets_tag:
        description: "Run only the assets job again for this release tag (project@version), from this branch"
        required: false
        default: ""
      assets_source:
        description: "With assets_tag: the code to build. current = the code of this branch as it is now, named with the tag's version (a fix reaches the files, but they carry newer code under an old number). tag = the tag's own commit, the exact code that was released. choose = stop and ask me"
        type: choice
        required: true
        default: choose
        options:
          - choose
          - current
          - tag
`
}

// githubCIOnlyWithoutAssetsTag is the ci job's condition when the workflow can
// be dispatched for a tag: that run builds files only, so the ci job skips.
func githubCIOnlyWithoutAssetsTag(m Model) string {
	if len(m.AssetPlatforms) == 0 {
		return ""
	}
	return "    if: ${{ inputs.assets_tag == '' }}\n"
}

// githubLinuxPackages is the step that installs [ci] native_linux_packages with
// apt on the Linux runners of the assets job, as the native job does, so the
// build command needs no apt call of its own. Empty when none are listed.
func githubLinuxPackages(packages []string) string {
	if len(packages) == 0 {
		return ""
	}
	var quoted []string
	for _, p := range packages {
		if strings.Trim(p, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.+:_-") != "" {
			p = "'" + strings.ReplaceAll(p, "'", `'\''`) + "'"
		}
		quoted = append(quoted, p)
	}
	return "      - name: Install the Linux packages\n        if: runner.os == 'Linux'\n        run: sudo apt-get update && sudo apt-get install -y --no-install-recommends " + strings.Join(quoted, " ") + "\n"
}

// githubAssetsJob runs smith ci assets on a runner of each platform after the
// ci job released: each runner builds and uploads its own files to the GitHub
// Release with gh release upload, so no artifact storage is used. A dispatch
// with assets_tag runs it without the ci job, for that existing tag.
func githubAssetsJob(m Model) string {
	if len(m.AssetPlatforms) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, `
  assets:
    needs: ci
    if: ${{ always() && ((needs.ci.outputs.released == 'true' && github.event_name == 'push' && github.ref_name == '%s') || (github.event_name == 'workflow_dispatch' && inputs.assets_tag != '')) }}
    strategy:
      fail-fast: false
      matrix:
        include:
`, m.ReleaseBranch)
	for _, p := range m.AssetPlatforms {
		fmt.Fprintf(&b, "          - platform: %s\n            os: %s\n", p, githubAssetRunners[p])
	}
	b.WriteString(`    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
`)
	b.WriteString(githubSiblingCheckouts(m))
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	for _, a := range m.Authenticate {
		b.WriteString(githubAuthenticate(a))
	}
	b.WriteString(githubInstallSmith(m, true))
	b.WriteString(githubSiblingPreBuilds(m))
	b.WriteString(githubLinuxPackages(m.NativeLinuxPackages))
	b.WriteString("      - name: Setup\n        run: smith ci setup --assets\n")
	b.WriteString(githubEnv("        ", m.SetupCredentials()))
	b.WriteString(`      - name: Build and upload the release files
        run: smith ci assets --tag "$ASSETS_TAG" --source "$ASSETS_SOURCE"
        env:
          ASSETS_TAG: ${{ inputs.assets_tag }}
          ASSETS_SOURCE: ${{ inputs.assets_source }}
          SMITH_PLATFORM: ${{ matrix.platform }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`)
	// The build command may read the secrets [ci] secrets names, and the cache's.
	for _, c := range m.TaskCredentials() {
		if c.Env != "GITHUB_TOKEN" {
			fmt.Fprintf(&b, "          %s: ${{ secrets.%s }}\n", c.Env, c.Env)
		}
	}
	return b.String()
}
