package release

// NothingToRelease is what to say when a release plan has no items: the
// reason, so the person can tell a workspace that is up to date from one whose
// commit messages do not count.
func NothingToRelease() string {
	return "Nothing to release: no package has a feat: or fix: commit since its last release (or since the start, if it was never released).\n" +
		"  → Releases come from conventional commits: commit with a message like \"feat: add search\" or \"fix: handle empty input\" (see Conventional Commits in docs/guides/getting-started.md), then run this again."
}
