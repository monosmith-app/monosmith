package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// GitLabPath is where the GitLab CI pipeline is written.
const GitLabPath = ".gitlab-ci.yml"

// GitLabImage is the image the job runs on; the toolchains install onto it.
const GitLabImage = "debian:bookworm"

// RenderGitLab renders the model as a GitLab CI pipeline: one job, run for
// merge requests and on the release branch, with the same phases as the
// other providers.
func RenderGitLab(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	fmt.Fprintf(&b, `
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
%[1]s
ci:
  image: %[2]s
  variables:
    GIT_DEPTH: "0"
    SMITH_VERSION: %[3]s
%[4]s%[5]s  before_script:
    - apt-get update -qq && apt-get install -y -qq curl git ca-certificates xz-utils >/dev/null
`, gitlabPushRules(m), GitLabImage, yamlValue(m.SmithVersion), gitlabStoreVariables(m.Stores), gitlabStores(m.Stores))
	for _, t := range m.Toolchains {
		b.WriteString(gitlabSetup(t))
	}
	install := SmithInstallFor(m)
	if install.Script {
		fmt.Fprintf(&b, "    - sh %s \"$SMITH_VERSION\"\n", InstallScriptPath)
	} else {
		b.WriteString("    - " + yamlValue(install.ContainerCommand()) + "\n")
		if dir := install.ContainerPathDir(); dir != "" {
			b.WriteString("    - " + yamlValue(`export PATH="`+dir+`:$PATH"`) + "\n")
		}
	}
	b.WriteString("    - export PATH=\"$HOME/.local/bin:$PATH\"\n")
	b.WriteString(slot("    ", "before-setup"))
	b.WriteString(`  script:
    - smith ci setup
    - smith ci audit
    - smith ci verify
`)
	b.WriteString(slot("    ", "after-verify"))
	b.WriteString("    # The release preview comments on the merge request with GITLAB_TOKEN (api scope).\n")
	b.WriteString("    - " + yamlValue(`if [ "$CI_PIPELINE_SOURCE" = "merge_request_event" ]; then smith release --dry-run --comment; fi`) + "\n")
	onPush := fmt.Sprintf(`[ "$CI_PIPELINE_SOURCE" = "push" ] && [ "$CI_COMMIT_BRANCH" = "%s" ]`, m.ReleaseBranch)
	if m.Pack {
		b.WriteString("    - " + yamlValue("if "+onPush+"; then smith ci pack; fi") + "\n")
	}
	onRelease := fmt.Sprintf(`[ "$CI_PIPELINE_SOURCE" = "push" ] && %s`, m.branchCondition(func(branch string) string { return fmt.Sprintf(`[ "$CI_COMMIT_BRANCH" = "%s" ]`, branch) }, func(c []string) string { return "{ " + strings.Join(c, " || ") + "; }" }))
	b.WriteString("    # Releasing pushes tags: GITLAB_TOKEN needs write_repository.\n")
	b.WriteString("    - " + yamlValue("if "+onRelease+`; then git remote set-url origin "https://oauth2:${GITLAB_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.git"; smith ci release; fi`) + "\n")
	if m.Pack {
		b.WriteString(`  artifacts:
    name: drop
    paths:
      - dist/drop
    when: on_success
    expire_in: 30 days
`)
	}
	if m.Native {
		b.WriteString("\n# Native projects build on every OS in the GitHub and Azure pipelines; on GitLab,\n# add a job per runner OS that runs `smith ci native` (it builds only the host's).\n")
	}
	if len(m.Credentials) > 0 {
		b.WriteString("\n# Secrets the release reads, as masked CI/CD variables:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "#   %s: %s\n", c.Env, c.Description)
		}
	}
	return b.String()
}

// gitlabSetup installs a toolchain on the job's Debian image with the
// plugin's install command, one script line per line.
func gitlabSetup(t protocol.Toolchain) string {
	if t.Install == "" {
		return missing("    ", t)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "    # %s %s\n", t.Name, t.Version)
	for _, line := range strings.Split(strings.TrimSpace(t.Install), "\n") {
		b.WriteString("    - " + yamlValue(strings.TrimSpace(line)) + "\n")
	}
	return b.String()
}

// gitlabPushRules are the workflow rules that run a pipeline on a push to the
// release branch and to each channel branch.
func gitlabPushRules(m Model) string {
	var b strings.Builder
	for _, branch := range m.ReleaseBranches() {
		fmt.Fprintf(&b, "    - if: $CI_COMMIT_BRANCH == \"%s\" && $CI_PIPELINE_SOURCE == \"push\"\n", branch)
	}
	return b.String()
}
