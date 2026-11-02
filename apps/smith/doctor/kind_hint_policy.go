package doctor

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// settledAfter is how long a lib has been in the repository before an
// importer-less lib says something: a lib made a minute ago, to be wired up in
// the next commit, is not a finding.
const settledAfter = 14 * 24 * time.Hour

// UnusedLibHint is the gentle suggestion for a lib no project imports: it is
// meant to be shared inside the repository, and it is shared with nobody.
// ok is false when there is nothing to say: the lib is used, it is the only
// project, or it was added too recently to judge (added is zero when git does
// not know, and then nothing is said).
func UnusedLibHint(name string, dependents, otherProjects int, added, now time.Time) (protocol.Finding, bool) {
	if dependents > 0 || otherProjects == 0 || added.IsZero() || now.Sub(added) < settledAfter {
		return protocol.Finding{}, false
	}
	days := int(now.Sub(added).Hours() / 24)
	return protocol.Finding{
		Check:   "the lib " + name + " is used by a project",
		Warning: true,
		Detail:  fmt.Sprintf("no project imports %s, and it has been in the repository for %d days. A lib is shared inside the repository; nothing shares it yet", name, days),
		Remedy:  "if it is a program, `smith kind " + name + " app`; if you publish it, `smith kind " + name + " package` (add --dry-run to see what changes first); if nothing needs it, `smith remove " + name + "`. Leave it if something outside the graph uses it",
		Fix: smithfix.Choice("Say what the lib "+name+" is", "What is "+name+"? You see what changes before anything is written.",
			smithfix.Option("A program (app): it is run, not imported", "kind", name, "app"),
			smithfix.Option("A package: it is published", "kind", name, "package")),
	}, true
}

// UnpublishedPackageHint is the gentle suggestion for a package a release
// never uploads while the workspace's other packages are: its own registry is
// none (smith.toml), or its manifest says it is never published. A package is
// what a release publishes.
func UnpublishedPackageHint(name string, neverUploaded bool) (protocol.Finding, bool) {
	if !neverUploaded {
		return protocol.Finding{}, false
	}
	return protocol.Finding{
		Check:   "the package " + name + " is published",
		Warning: true,
		Detail:  name + " is a package, but nothing publishes it: its own settings say registry is none, or its manifest says it is never published, so a release only tags it",
		Remedy:  "if it only serves other projects of this repository, `smith kind " + name + " lib` says what changes (add --dry-run to see first); if you publish it another way, leave it",
		Fix:     smithfix.Fix("Make the package "+name+" a lib", smithfix.Step("Make "+name+" a lib: it is shared inside the repository and no longer published", "kind", name, "lib")),
	}, true
}
