package cipipeline

import (
	"fmt"
	"strings"
)

// manualReleaseScript is the shell that runs the release: with the value a
// person gave it passes it to smith ci release, else it is a normal release.
// The value comes from an environment variable, never from the script text.
const manualReleaseScript = `if [ -n "$RELEASE_BUMP" ]; then smith ci release --bump "$RELEASE_BUMP" --force; else smith ci release; fi`

// azureParameters is the runtime parameter a person fills in when they run the
// Azure pipeline by hand. Empty unless [ci] manual_release is on.
func azureParameters(m Model) string {
	if !m.ManualRelease {
		return ""
	}
	return `parameters:
  - name: releaseBump
    displayName: "Release by hand a package no commit releases: project=patch|minor|major (publishes and cannot be undone; empty: no manual release)"
    type: string
    default: ""

`
}

// azureManualCondition is the condition of a run started by hand with a value
// in releaseBump, on a release branch.
func azureManualCondition(m Model) string {
	branches := m.branchCondition(
		func(branch string) string {
			return fmt.Sprintf("eq(variables['Build.SourceBranch'], 'refs/heads/%s')", branch)
		},
		func(c []string) string { return "or(" + strings.Join(c, ", ") + ")" })
	return "and(succeeded(), eq(variables['Build.Reason'], 'Manual'), ne('${{ parameters.releaseBump }}', ''), " + branches + ")"
}

// azureReleaseStep is the Release step: it runs after a push to a release
// branch and, with [ci] manual_release, after a run started by hand with a
// releaseBump.
func azureReleaseStep(m Model, onRelease string) string {
	var b strings.Builder
	if !m.ManualRelease {
		fmt.Fprintf(&b, "  - bash: smith ci release\n    displayName: Release\n    condition: %s\n", onRelease)
		if len(m.Credentials) > 0 {
			b.WriteString("    env:\n")
		}
	} else {
		fmt.Fprintf(&b, "  - bash: %s\n    displayName: Release\n    condition: or(%s, %s)\n    env:\n      RELEASE_BUMP: ${{ parameters.releaseBump }}\n",
			yamlValue(manualReleaseScript), onRelease, azureManualCondition(m))
	}
	for _, c := range m.Credentials {
		fmt.Fprintf(&b, "      %s: $(%s)\n", c.Env, c.Env)
	}
	return b.String()
}

// gitlabVariables is the top-level variable that the "Run pipeline" form of
// GitLab shows, prefilled with its description. Empty unless [ci]
// manual_release is on.
func gitlabVariables(m Model) string {
	if !m.ManualRelease {
		return ""
	}
	return `
variables:
  RELEASE_BUMP:
    value: ""
    description: "Release by hand a package no commit releases: project=patch|minor|major (publishes and cannot be undone; empty: no manual release)"
`
}

// gitlabManualRule lets a pipeline started from the web form run.
func gitlabManualRule(m Model) string {
	if !m.ManualRelease {
		return ""
	}
	return "    - if: $CI_PIPELINE_SOURCE == \"web\" && $RELEASE_BUMP != \"\"\n"
}

// gitlabRelease is the release script line: after a push to a release branch,
// or, with [ci] manual_release, after a run from the web form with a value.
func gitlabRelease(m Model, onRelease string) string {
	command := `if ` + onRelease + `; then git remote set-url origin "https://oauth2:${GITLAB_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.git"; smith ci release; fi`
	if !m.ManualRelease {
		return "    - " + yamlValue(command) + "\n"
	}
	onWeb := `[ "$CI_PIPELINE_SOURCE" = "web" ] && ` + m.branchCondition(func(branch string) string { return fmt.Sprintf(`[ "$CI_COMMIT_BRANCH" = "%s" ]`, branch) }, func(c []string) string { return "{ " + strings.Join(c, " || ") + "; }" })
	return "    - " + yamlValue(`if `+onRelease+` || `+onWeb+`; then git remote set-url origin "https://oauth2:${GITLAB_TOKEN}@${CI_SERVER_HOST}/${CI_PROJECT_PATH}.git"; `+manualReleaseScript+`; fi`) + "\n"
}

// bitbucketCustomPipeline is the pipeline a person runs by hand from the
// "Run pipeline" form, which asks for RELEASE_BUMP. Empty unless [ci]
// manual_release is on.
func bitbucketCustomPipeline(m Model, steps string) string {
	if !m.ManualRelease {
		return ""
	}
	return "  custom:\n    release-by-hand:\n      - variables:\n          - name: RELEASE_BUMP\n            description: \"Release by hand a package no commit releases: project=patch|minor|major (publishes and cannot be undone)\"\n" + steps
}

// circleciManualReleaseNote says that CircleCI has no form for a manual value.
const circleciManualReleaseNote = "CircleCI: [ci] manual_release does not apply; CircleCI has no form to run a pipeline by hand with a value, so release a package no commit releases with `smith release --bump <project>=patch --force` on a machine, or from the GitHub, Azure, GitLab or Bitbucket pipeline"
