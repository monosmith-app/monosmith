package release

import "strings"

// FirstReleaseCommand is the commit that makes a first release, shown to the
// person and never made by smith.
const FirstReleaseCommand = `git commit --allow-empty -m "feat: first version"`

// NothingToRelease is what to say when a release plan has no items. It tells a
// workspace that has no package (apps and internal libraries never release), a
// workspace whose packages are all held back, and a workspace that is up to date
// apart. Only a workspace that has a package names the commit types that count
// and shows the exact command for a first release, which smith does not run.
//
// Parameters: plan is the plan after --hold and --bump; Held and Unchanged say
// which case it is.
// Returns: the text, one line per sentence.
func NothingToRelease(plan Plan) string {
	switch {
	case len(plan.Held) > 0:
		return "Nothing to release: you held back " + strings.Join(plan.Held, ", ") + ", and no other package has a feat: or fix: commit.\n" +
			"  → Run smith release again without --hold to release " + holdsPhrase(plan.Held) + "."
	case len(plan.Unchanged) == 0 && plan.Channel == "":
		return "Nothing to release: this workspace has no package. Apps and internal libraries are never released.\n" +
			"  → Add a package with smith add package <name> --lang <language>.\n" +
			"  → Or turn a project that other projects import into a package with smith kind <name> package. It is then published to a registry."
	}
	return "Nothing to release: no package has a feat: or fix: commit since its last release (or since the start, if it was never released).\n" +
		"  → Only feat: (a minor version) and fix: (a patch) commits make a release. The commits smith makes itself, such as \"chore: add app web\", never do.\n" +
		"  → A project releases when a feat:, fix: or breaking commit touches its own folder (or a lib it imports). A chore: commit, or a change only under tools/ or in smith.toml, releases nothing.\n" +
		"  → Commit your own work with a message like \"feat: add search\" or \"fix: handle empty input\" (see https://github.com/monosmith-app/monosmith/blob/main/docs/getting-started/first-workspace.md), then run this again.\n" +
		"  → For a first release of packages that were never released, this is the commit that counts (smith does not make it for you):\n" +
		"      " + FirstReleaseCommand
}

func holdsPhrase(held []string) string {
	if len(held) == 1 {
		return held[0]
	}
	return "them"
}
