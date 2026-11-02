package release

// FirstReleaseCommand is the commit that makes a first release, shown to the
// person and never made by smith.
const FirstReleaseCommand = `git commit --allow-empty -m "feat: first version"`

// NothingToRelease is what to say when a release plan has no items: the
// reason, so the person can tell a workspace that is up to date from one whose
// commit messages do not count. It names the commit types that count, says that
// the commits smith makes itself (chore: add app web) never count, and shows the
// exact command for a first release, which smith does not run.
func NothingToRelease() string {
	return "Nothing to release: no package has a feat: or fix: commit since its last release (or since the start, if it was never released).\n" +
		"  → Only feat: (a minor version) and fix: (a patch) commits make a release. The commits smith makes itself, such as \"chore: add app web\", never do.\n" +
		"  → Commit your own work with a message like \"feat: add search\" or \"fix: handle empty input\" (see https://github.com/monosmith-app/monosmith/blob/main/docs/getting-started/first-workspace.md), then run this again.\n" +
		"  → For a first release of projects that were never released, this is the commit that counts (smith does not make it for you):\n" +
		"      " + FirstReleaseCommand
}
