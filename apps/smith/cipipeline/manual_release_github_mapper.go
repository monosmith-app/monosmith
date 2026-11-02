package cipipeline

// githubReleaseBumpInput is the workflow_dispatch input that starts a manual
// release (see ManualRelease). Empty unless [ci] manual_release is on.
func githubReleaseBumpInput(m Model) string {
	if !m.ManualRelease {
		return ""
	}
	return `      release_bump:
        description: "Release by hand a package no commit releases, as project=patch|minor|major (several: separate with commas). This publishes to the registries, tags HEAD and cannot be undone. Empty: no manual release"
        required: false
        default: ""
`
}

// githubReleaseEvent is the condition that holds when a run may release: a push,
// or with [ci] manual_release a dispatch that carries a release_bump.
func githubReleaseEvent(m Model) string {
	if !m.ManualRelease {
		return "github.event_name == 'push'"
	}
	return "(github.event_name == 'push' || (github.event_name == 'workflow_dispatch' && inputs.release_bump != ''))"
}

// githubReleaseRun is the release step's command: with [ci] manual_release it
// passes the input to smith ci release, read from the environment so that the
// input value is never part of the script.
func githubReleaseRun(m Model) string {
	if !m.ManualRelease {
		return "        run: smith ci release\n"
	}
	return `        run: |
          if [ -n "$RELEASE_BUMP" ]; then
            smith ci release --bump "$RELEASE_BUMP" --force
          else
            smith ci release
          fi
`
}

// githubReleaseBumpEnv is the environment line that hands the input to the step.
func githubReleaseBumpEnv(m Model) string {
	if !m.ManualRelease {
		return ""
	}
	return "          RELEASE_BUMP: ${{ inputs.release_bump }}\n"
}
