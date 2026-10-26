package adoptwizard

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
)

// The review paints lines that start with + green and - red (they are
// diffs), so the text of a screen never starts a line with either: lists use
// a bullet.

// shownProjects is how many projects the first screen names.
const shownProjects = 12

// shownKept is how many kept files and packages the last screen names.
const shownKept = 8

// shownChanged is how many changes of behaviour the last screen names.
const shownChanged = 5

// foundText is the first screen: what the folder is and what it uses.
func foundText(f adoptflow.Found) string {
	var b strings.Builder
	switch f.Kind {
	case adoptflow.KindSmith, adoptflow.KindNotARepo:
		fmt.Fprintf(&b, "● %s\n", f.Name)
	default:
		fmt.Fprintf(&b, "● Found %s %s: %s\n", article(f.Source), f.Source, f.Name)
		if len(f.Blockers) == 0 {
			fmt.Fprintf(&b, "  %s\n", projectsLine(f))
		}
		if len(f.Tooling) > 0 {
			fmt.Fprintf(&b, "  Tools it uses today: %s\n", strings.Join(f.Tooling, ", "))
		} else if len(f.Blockers) == 0 {
			fmt.Fprintf(&b, "  Nothing to replace: it has no lint, test or build tool of its own.\n")
		}
	}
	for i, p := range f.Projects {
		if i == shownProjects {
			fmt.Fprintf(&b, "  • and %d more\n", len(f.Projects)-shownProjects)
			break
		}
		fmt.Fprintf(&b, "  • %s  %s  %s\n", p.Name, projectWords(p), p.Root)
	}
	for _, x := range f.Blockers {
		fmt.Fprintf(&b, "\n✖ %s\n", x)
	}
	for _, x := range f.Warnings {
		fmt.Fprintf(&b, "\n! %s\n", x)
	}
	switch {
	case f.NeedsInstall:
		fmt.Fprintf(&b, "\nNothing has changed. Installing only fills node_modules; git does not track it, so git cannot take it back.\n")
	case len(f.Blockers) == 0:
		fmt.Fprintf(&b, "\nNothing has changed. The next screens ask how far to go and list every file before anything is written.\n")
		if !f.Full && f.NoFull != "" {
			fmt.Fprintf(&b, "\n%s\n", f.NoFull)
		}
	default:
		fmt.Fprintf(&b, "\nNothing was changed.\n")
	}
	return b.String()
}

// article is "a" or "an" for a noun phrase as it is said aloud ("an Nx
// workspace", "a Turborepo workspace").
func article(noun string) string {
	if strings.HasPrefix(noun, "Nx") || strings.ContainsRune("AEIOUaeiou", rune(noun[0])) {
		return "an"
	}
	return "a"
}

// projectWords says what a project is: "TypeScript / JavaScript library".
func projectWords(p adoptflow.Project) string {
	return strings.TrimSpace(languagenames.Of(p.Language) + " " + languagenames.KindOf(p.Kind))
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
func conservativeText(f adoptflow.Found, commandReview string, changes adoptflow.Changes) string {
	var b strings.Builder
	if f.Kind.Convertible() {
		fmt.Fprintf(&b, "● Convert the %s %q to smith, keeping what works\n", f.Source, f.Name)
		fmt.Fprintf(&b, "  Your lint, test and build tools stay as they are. smith adds its settings file (smith.toml),\n")
		fmt.Fprintf(&b, "  updates each project's package.json, and replaces the CI pipeline files with ones it generates\n")
		fmt.Fprintf(&b, "  (your own steps go in the marked sections). What needs a look afterwards is listed at the end.\n")
		fmt.Fprintf(&b, "  Every change is a file in git, so git can take it back.\n\n")
		b.WriteString(changes.Text())
		return b.String()
	}
	b.WriteString(commandReview)
	fmt.Fprintf(&b, "\n  Writes smith.toml and the CI pipelines for the %d projects found and makes one commit;\n", len(f.Projects))
	fmt.Fprintf(&b, "  your build and lint setup stay as they are. smith doctor then says what is left to do.\n")
	return b.String()
}

// reviewHeader goes above the full plan.
const reviewHeader = "● Nothing has changed yet. The files below are what the conversion writes. After it, smith asks about each group of old files before deleting anything. Every change is a file in git, so git can take it back.\n\n"

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

// doneText is the last screen, in three parts: what is done, what behaves
// differently now, and what needs the person.
func doneText(f adoptflow.Found, r convert.Highlights, applied, declined []string, kept []legacyinventory.Kept, notes []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "✓ %q is a smith workspace now", f.Name)
	if r.Projects != "" {
		fmt.Fprintf(&b, ": %s", r.Projects)
	}
	b.WriteString("\n")
	if len(applied) > 0 {
		fmt.Fprintf(&b, "  Old tools removed: %s\n", strings.Join(applied, ", "))
	}
	if len(declined) > 0 {
		fmt.Fprintf(&b, "  Left as they were, as you chose: %s\n", strings.Join(declined, ", "))
	}
	fmt.Fprintf(&b, "  Nothing is committed yet: git status lists the changes, and git can take them back.\n")
	if len(r.Checks) > 0 {
		fmt.Fprintf(&b, "\n! Needs you (%d)\n", len(r.Checks))
		for _, c := range r.Checks {
			fmt.Fprintf(&b, "  • %s\n", c)
		}
	}
	for _, n := range notes {
		fmt.Fprintf(&b, "\n! %s\n", n)
	}
	if len(r.Changed) > 0 {
		fmt.Fprintf(&b, "\n● Behaves differently now (%d)\n", len(r.Changed))
		for i, c := range r.Changed {
			if i == shownChanged {
				fmt.Fprintf(&b, "  • and %d more: the full list is in %s\n", len(r.Changed)-shownChanged, convert.ReportFile)
				break
			}
			fmt.Fprintf(&b, "  • %s\n", c)
		}
	}
	if len(kept) > 0 {
		fmt.Fprintf(&b, "\nLooks like the old tooling but stays (something still uses it):\n")
		for i, k := range kept {
			if i == shownKept {
				fmt.Fprintf(&b, "  • and %d more\n", len(kept)-shownKept)
				break
			}
			fmt.Fprintf(&b, "  • %s: %s\n", k.What, k.Why)
		}
	}
	fmt.Fprintf(&b, "\nOpen the workspace to see it, or let smith build everything first to check that the conversion works.\n")
	return b.String()
}
