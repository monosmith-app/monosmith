package adoptwizard

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

// The review paints lines that start with + green and - red (they are
// diffs), so the text of a screen never starts a line with either: lists use
// a bullet.

// shownProjects is how many projects the first screen names.
const shownProjects = 12

// shownKept is how many kept files and packages the last screen names.
const shownKept = 8

// foundText is the first screen: what the folder is and what it uses.
func foundText(f adoptflow.Found) string {
	var b strings.Builder
	switch f.Kind {
	case adoptflow.KindSmith, adoptflow.KindNotARepo:
		fmt.Fprintf(&b, "● %s\n", f.Name)
	default:
		fmt.Fprintf(&b, "● Found a %s: %s\n", f.Source, f.Name)
		fmt.Fprintf(&b, "  %s\n", projectsLine(f))
		if len(f.Tooling) > 0 {
			fmt.Fprintf(&b, "  Old tooling in use: %s\n", strings.Join(f.Tooling, ", "))
		} else {
			fmt.Fprintf(&b, "  No old lint, test or build tooling found to replace.\n")
		}
	}
	for i, p := range f.Projects {
		if i == shownProjects {
			fmt.Fprintf(&b, "  • and %d more\n", len(f.Projects)-shownProjects)
			break
		}
		fmt.Fprintf(&b, "  • %s  %s  %s\n", p.Name, strings.TrimSpace(p.Language+" "+p.Kind), p.Root)
	}
	for _, x := range f.Blockers {
		fmt.Fprintf(&b, "\n✖ %s\n", x)
	}
	for _, x := range f.Warnings {
		fmt.Fprintf(&b, "\n! %s\n", x)
	}
	if len(f.Blockers) == 0 {
		fmt.Fprintf(&b, "\nNothing has changed. The next screens ask how far to go and show every change before anything is written.\n")
		if !f.Full && f.NoFull != "" {
			fmt.Fprintf(&b, "\n%s\n", f.NoFull)
		}
	} else {
		fmt.Fprintf(&b, "\nNothing was changed.\n")
	}
	return b.String()
}

func projectsLine(f adoptflow.Found) string {
	switch len(f.Projects) {
	case 0:
		return "No projects found yet."
	case 1:
		return "1 project:"
	}
	return fmt.Sprintf("%d projects:", len(f.Projects))
}

// conservativeText is the review of the conservative level.
func conservativeText(f adoptflow.Found, commandReview string) string {
	var b strings.Builder
	if f.Kind.Convertible() {
		fmt.Fprintf(&b, "● Convert the %s %q to smith, keeping what works\n", f.Source, f.Name)
		fmt.Fprintf(&b, "  Writes smith.toml, edits each project's package.json, takes out the old tool's files,\n")
		fmt.Fprintf(&b, "  regenerates the CI pipelines, then installs the packages and refreshes package-lock.json.\n")
		fmt.Fprintf(&b, "  The old lint, test and build tooling stays, and the report lists what to check.\n")
		fmt.Fprintf(&b, "  smith convert --dry-run lists every file. git undoes it all.\n")
		return b.String()
	}
	b.WriteString(commandReview)
	fmt.Fprintf(&b, "\n  Writes smith.toml and the CI pipelines for the %d projects found and makes one commit;\n", len(f.Projects))
	fmt.Fprintf(&b, "  your build and lint setup stay as they are. smith doctor then says what is left to do.\n")
	return b.String()
}

// reviewHeader goes above the full plan.
const reviewHeader = "● Nothing has changed yet. Enter converts the workspace, then asks about each group below before it deletes anything; git undoes it all.\n\n"

// groupText is the question about one group of removals, above its diffs.
func groupText(g removalplan.Group, reviewed string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "● We are going to delete %s. %s — ok?\n", groupCounts(g), g.Reason)
	if len(g.Removes) > 0 {
		fmt.Fprintf(&b, "  packages: %s\n", strings.Join(g.Removes, ", "))
	}
	b.WriteString("  y deletes them · n keeps them · s keeps these and skips the groups after\n\n")
	b.WriteString(reviewed)
	return b.String()
}

func groupCounts(g removalplan.Group) string {
	var parts []string
	if g.Packages > 0 {
		parts = append(parts, plural(g.Packages, "package"))
	}
	if g.Files > 0 {
		parts = append(parts, plural(g.Files, "file"))
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, " and ")
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// doneText is the last screen: what was done and what stays.
func doneText(f adoptflow.Found, summary string, applied, declined []string, kept []legacyinventory.Kept, notes []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "● %q is a smith workspace now\n", f.Name)
	if s := strings.TrimSpace(summary); s != "" {
		fmt.Fprintf(&b, "\n%s\n", s)
	}
	if len(applied) > 0 {
		fmt.Fprintf(&b, "\nOld tooling removed: %s\n", strings.Join(applied, ", "))
	}
	if len(declined) > 0 {
		fmt.Fprintf(&b, "\nLeft as it was, as you chose: %s\n", strings.Join(declined, ", "))
	}
	if len(kept) > 0 {
		fmt.Fprintf(&b, "\nLooks like the old tooling but stays (something still uses it):\n")
		for i, k := range kept {
			if i == shownKept {
				fmt.Fprintf(&b, "  • and %d more (smith convert --level full --json lists them)\n", len(kept)-shownKept)
				break
			}
			fmt.Fprintf(&b, "  • %s: %s\n", k.What, k.Why)
		}
	}
	for _, n := range notes {
		fmt.Fprintf(&b, "\n! %s\n", n)
	}
	fmt.Fprintf(&b, "\nEnter closes this; run smith again to open the workspace.\n")
	return b.String()
}
