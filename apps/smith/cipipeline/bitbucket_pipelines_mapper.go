package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// BitbucketPath is where the Bitbucket Pipelines configuration is written.
const BitbucketPath = "bitbucket-pipelines.yml"

// BitbucketImage is the image the steps run on; the toolchains install onto
// it with the plugins' shell commands, as on GitLab CI.
const BitbucketImage = "debian:bookworm"

// bitbucketCacheFiles is how many lockfiles a Bitbucket cache key can hash.
const bitbucketCacheFiles = 10

// RenderBitbucket renders the model as a Bitbucket Pipelines configuration:
// one `ci` step with the same phases as the other providers, a `native` step
// after it when a project is native, both defined once and run for pull
// requests and on the release branch.
//
// A step's script runs in one shell, so the toolchains' commands (written for
// GitLab) work unchanged. The release steps test the pull request variable at
// run time, so the one step serves both pipelines.
func RenderBitbucket(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	fmt.Fprintf(&b, "\nimage: %s\n\nclone:\n  depth: full\n\ndefinitions:\n", BitbucketImage)
	b.WriteString(bitbucketCaches(m.Stores))
	b.WriteString("  steps:\n    - step: &ci\n        name: Verify\n")
	b.WriteString(bitbucketCacheList(m.Stores))
	b.WriteString("        script:\n")
	b.WriteString(bitbucketInstall(m))
	b.WriteString(slot("          ", "before-setup"))
	b.WriteString("          - smith ci setup\n          - smith ci audit\n          - smith ci verify\n")
	b.WriteString(slot("          ", "after-verify"))
	b.WriteString("          # The release preview comments on the pull request with BITBUCKET_TOKEN (a repository access token\n")
	b.WriteString("          # with pull request write).\n")
	b.WriteString("          - " + yamlValue(`if [ -n "$BITBUCKET_PR_ID" ]; then smith release --dry-run --comment; fi`) + "\n")
	onRelease := fmt.Sprintf(`[ -z "$BITBUCKET_PR_ID" ] && [ "$BITBUCKET_BRANCH" = "%s" ]`, m.ReleaseBranch)
	if m.Pack {
		b.WriteString("          - " + yamlValue("if "+onRelease+"; then smith ci pack; fi") + "\n")
	}
	b.WriteString("          # Releasing pushes tags: BITBUCKET_TOKEN needs repository write.\n")
	b.WriteString("          - " + yamlValue("if "+onRelease+`; then git remote set-url origin "https://x-token-auth:${BITBUCKET_TOKEN}@bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}.git"; smith ci release; fi`) + "\n")
	if m.Pack {
		b.WriteString("        artifacts:\n          - dist/drop/**\n")
	}
	if m.Native {
		b.WriteString(bitbucketNativeStep(m))
	}

	steps := "      - step: *ci\n"
	if bitbucketNative(m) {
		steps += "      - step: *native\n"
	}
	fmt.Fprintf(&b, "\npipelines:\n  pull-requests:\n    '**':\n%s  branches:\n    %s:\n%s", steps, yamlWord(m.ReleaseBranch), steps)
	if len(m.Credentials) > 0 {
		b.WriteString("\n# Secrets the release reads, as secured repository or workspace variables:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "#   %s: %s\n", c.Env, c.Description)
		}
	}
	return b.String()
}

// bitbucketNative is true when the native step is written: a project is
// native and a runner is Linux, the only system Bitbucket's cloud runs.
func bitbucketNative(m Model) bool {
	return m.Native && len(circleciLinuxRunners(m)) > 0
}

// bitbucketNativeStep builds the native projects after the ci step. Other
// systems need self-hosted runners, which are the team's to name, so the
// skipped runners are said in a comment.
func bitbucketNativeStep(m Model) string {
	var b strings.Builder
	var skipped []string
	for _, r := range m.NativeRunners {
		if !isLinuxRunner(r) {
			skipped = append(skipped, r)
		}
	}
	if len(skipped) > 0 {
		fmt.Fprintf(&b, "    # Not built on Bitbucket: %s. The cloud runs Linux; add a step with `runs-on: [self.hosted, ...]`\n    # on a self-hosted macOS or Windows runner that runs `smith ci native`.\n", strings.Join(skipped, ", "))
	}
	if !bitbucketNative(m) {
		return b.String()
	}
	b.WriteString("    - step: &native\n        name: Native (Linux)\n")
	b.WriteString(bitbucketCacheList(m.Stores))
	b.WriteString("        script:\n")
	b.WriteString(bitbucketInstall(m))
	b.WriteString("          - smith ci setup\n          - smith ci native\n")
	b.WriteString("        artifacts:\n          - dist/drop/**\n")
	return b.String()
}

// bitbucketInstall is the script lines that install the system packages, the
// toolchains and smith.
func bitbucketInstall(m Model) string {
	var b strings.Builder
	b.WriteString("          - apt-get update -qq && apt-get install -y -qq curl git ca-certificates xz-utils >/dev/null\n")
	for _, t := range m.Toolchains {
		if t.Install == "" {
			b.WriteString(missing("          ", t))
			continue
		}
		fmt.Fprintf(&b, "          # %s %s\n", t.Name, t.Version)
		for _, line := range strings.Split(strings.TrimSpace(t.Install), "\n") {
			b.WriteString("          - " + yamlValue(strings.TrimSpace(line)) + "\n")
		}
	}
	install := SmithInstallFor(m)
	if install.Script {
		b.WriteString("          - " + yamlValue("sh "+InstallScriptPath+" "+shellWord(m.SmithVersion)) + "\n")
	} else {
		b.WriteString("          - " + yamlValue(install.ContainerCommand()) + "\n")
		if dir := install.ContainerPathDir(); dir != "" {
			b.WriteString("          - " + yamlValue(`export PATH="`+dir+`:$PATH"`) + "\n")
		}
	}
	b.WriteString("          - " + yamlValue(`export PATH="$HOME/.local/bin:$PATH"`) + "\n")
	return b.String()
}

// bitbucketCaches defines one cache per package store, keyed on its lockfiles.
func bitbucketCaches(stores []protocol.PackageStore) string {
	if len(stores) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("  caches:\n")
	for _, s := range stores {
		files := s.Files
		if len(files) > bitbucketCacheFiles {
			files = files[:bitbucketCacheFiles]
		}
		fmt.Fprintf(&b, "    %s:\n      key:\n        files:\n", s.Name)
		for _, f := range files {
			fmt.Fprintf(&b, "          - %s\n", yamlWord(f))
		}
		fmt.Fprintf(&b, "      path: %s\n", s.Path)
	}
	return b.String()
}

// bitbucketCacheList is a step's caches key, one entry per store.
func bitbucketCacheList(stores []protocol.PackageStore) string {
	if len(stores) == 0 {
		return ""
	}
	names := make([]string, len(stores))
	for i, s := range stores {
		names[i] = s.Name
	}
	return "        caches: [" + strings.Join(names, ", ") + "]\n"
}
