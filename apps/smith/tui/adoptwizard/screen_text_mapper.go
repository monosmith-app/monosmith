package adoptwizard

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/libs/words/words"
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

// say writes the catalog text for key, with its placeholders filled from pairs.
func say(b *strings.Builder, key string, pairs ...any) { b.WriteString(words.Say(key, pairs...)) }

// foundText is the first screen: what the folder is and what it uses.
func foundText(f adoptflow.Found) string {
	var b strings.Builder
	switch f.Kind {
	case adoptflow.KindSmith, adoptflow.KindNotARepo:
		say(&b, "adopt.found.name", "name", f.Name)
	default:
		say(&b, "adopt.found.is", "article", article(f.Source), "source", f.Source, "name", f.Name)
		if len(f.Blockers) == 0 {
			b.WriteString("  " + projectsLine(f) + "\n")
		}
		if len(f.Tooling) > 0 {
			say(&b, "adopt.found.tooling", "tools", strings.Join(f.Tooling, ", "))
		} else if len(f.Blockers) == 0 {
			say(&b, "adopt.found.notooling")
		}
	}
	for i, p := range f.Projects {
		if i == shownProjects {
			say(&b, "adopt.more", "count", len(f.Projects)-shownProjects)
			break
		}
		b.WriteString("  • " + p.Name + "  " + projectWords(p) + "  " + p.Root + "\n")
	}
	if len(f.Blockers) > 0 && !f.NeedsInstall {
		say(&b, "adopt.found.blockers", "blockers", words.Count("adopt.blocker", len(f.Blockers)))
	}
	for _, x := range f.Blockers {
		// Missing packages are a step smith takes itself, not a dead end.
		mark := marks.Failed
		if f.NeedsInstall {
			mark = "●"
		}
		b.WriteString("\n" + mark + " " + x + "\n")
	}
	for _, x := range f.Warnings {
		b.WriteString("\n! " + x + "\n")
	}
	switch {
	case f.NeedsInstall:
		say(&b, "adopt.found.install")
	case len(f.Blockers) == 0 && f.RootProject != nil:
		say(&b, "adopt.found.root", "manifest", f.RootProject.Manifest, "name", strconv.Quote(f.RootProject.Name))
		say(&b, "adopt.found.unchanged.root")
	case len(f.Blockers) == 0:
		say(&b, "adopt.found.unchanged")
		if !f.Full && f.NoFull != "" {
			b.WriteString("\n" + f.NoFull + "\n")
		}
	default:
		say(&b, "adopt.found.blocked")
	}
	return b.String()
}

// article is "a" or "an" for a noun phrase as it is said aloud ("an Nx
// workspace", "a Turborepo workspace").
func article(noun string) string {
	if strings.HasPrefix(noun, "Nx") || strings.ContainsRune("AEIOUaeiou", rune(noun[0])) { // words:literal
		return words.Say("adopt.article.an")
	}
	return words.Say("adopt.article.a")
}

// projectWords says what a project is: "TypeScript / JavaScript library".
func projectWords(p adoptflow.Project) string {
	return strings.TrimSpace(languagenames.Of(p.Language) + " " + languagenames.KindOf(p.Kind))
}

func projectsLine(f adoptflow.Found) string {
	if len(f.Projects) == 0 {
		return words.Say("adopt.projects.none")
	}
	return words.Count("adopt.projects", len(f.Projects))
}

// conservativeText is the review of the conservative level.
func conservativeText(f adoptflow.Found, commandReview string, changes adoptflow.Changes, diffs string, r convert.Highlights, commit bool) string {
	var b strings.Builder
	if f.Kind.Convertible() {
		tool := strings.TrimSuffix(f.Source, " workspace") // words:literal
		say(&b, "adopt.convert.title", "source", f.Source, "name", strconv.Quote(f.Name))
		say(&b, "adopt.convert.body", "tool", tool)
		b.WriteString(planSummaryText(changes, r))
		if changes.Total() > 0 {
			// The same diffs as the Doctor fix dialog shows for a fix.
			b.WriteString("\n" + diffs)
		}
		b.WriteString(commitBox(commit))
		return b.String()
	}
	b.WriteString(commandReview)
	if len(f.Projects) == 0 && f.RootProject != nil {
		say(&b, "adopt.writes.moved")
	} else {
		say(&b, "adopt.writes.found", "count", len(f.Projects))
	}
	say(&b, "adopt.writes.kept")
	b.WriteString(commitBox(commit))
	return b.String()
}

// commitBox is the tick box that asks for the change to be committed; a click
// on its line, or the c key, flips it.
func commitBox(commit bool) string {
	return words.Say("adopt.commit.box", "box", choicerow.Box(commit, false))
}

// reviewHeader goes above the full plan.
var reviewHeader = words.Say("adopt.review.header")

// groupText is the question about one group of removals, above its diffs.
func groupText(g removalplan.Group, reviewed string) string {
	var b strings.Builder
	if g.Files == 0 && g.Packages == 0 {
		say(&b, "adopt.group.empty", "reason", g.Reason)
		say(&b, "adopt.group.empty.keys")
		b.WriteString(reviewed)
		return b.String()
	}
	say(&b, "adopt.group.delete", "counts", groupCounts(g), "reason", g.Reason)
	if len(g.Removes) > 0 {
		say(&b, "adopt.group.packages", "list", strings.Join(g.Removes, ", "))
	}
	say(&b, "adopt.group.keys")
	b.WriteString(reviewed)
	return b.String()
}

func groupCounts(g removalplan.Group) string {
	var parts []string
	if g.Packages > 0 {
		parts = append(parts, words.Count("adopt.package", g.Packages))
	}
	if g.Files > 0 {
		parts = append(parts, words.Count("adopt.file", g.Files))
	}
	if len(parts) == 0 {
		return words.Say("adopt.nothing")
	}
	return strings.Join(parts, words.Say("adopt.and"))
}

// doneText is the last screen, in three parts: what is done, what behaves
// differently now, and what needs the person (the report's checks and what the
// closing checks found: findings, said with their remedies).
func doneText(f adoptflow.Found, r convert.Highlights, checked string, findings []doctorfix.Finding, applied, declined []string, kept []legacyinventory.Kept, notes []string, committed bool) string {
	var b strings.Builder
	if len(f.Projects) == 0 && r.Projects == "" {
		say(&b, "adopt.done.empty", "name", strconv.Quote(f.Name))
		say(&b, "adopt.done.empty.hint")
	} else {
		say(&b, "adopt.done.workspace", "name", strconv.Quote(f.Name))
		if r.Projects != "" {
			b.WriteString(": " + r.Projects)
		}
		b.WriteString("\n")
	}
	if checked != "" {
		say(&b, "adopt.done.works", "checked", checked)
		// A closing check that found something (the lint, doctor) needs the
		// person as much as a check of the report does.
		switch needs := len(r.Checks) + len(findings); needs {
		case 0:
			say(&b, "adopt.done.needs.none")
		default:
			b.WriteString(words.Count("adopt.done.needs", needs))
		}
	}
	if len(applied) > 0 {
		say(&b, "adopt.done.removed", "list", strings.Join(applied, ", "))
	}
	if len(declined) > 0 {
		say(&b, "adopt.done.declined", "list", strings.Join(declined, ", "))
	}
	if committed {
		say(&b, "adopt.done.committed")
	} else {
		say(&b, "adopt.done.uncommitted")
	}
	if len(r.Checks) > 0 {
		say(&b, "adopt.done.checks", "count", len(r.Checks))
		for _, c := range r.Checks {
			b.WriteString("  • " + c + "\n")
		}
	}
	if len(findings) > 0 {
		say(&b, "adopt.done.findings", "count", len(findings))
		for _, f := range findings {
			for _, line := range strings.SplitAfter(doctorfix.FindingText(f), "\n") {
				if line != "" {
					b.WriteString("  " + line)
				}
			}
		}
	}
	for _, n := range notes {
		b.WriteString("\n! " + n + "\n")
	}
	if len(r.Changed) > 0 {
		say(&b, "adopt.done.changed", "count", len(r.Changed))
		for i, c := range r.Changed {
			if i == shownChanged {
				say(&b, "adopt.done.changed.more", "count", len(r.Changed)-shownChanged, "file", convert.ReportFile)
				break
			}
			b.WriteString("  • " + c + "\n")
		}
	}
	if len(kept) > 0 {
		say(&b, "adopt.done.kept")
		for i, k := range kept {
			if i == shownKept {
				say(&b, "adopt.more", "count", len(kept)-shownKept)
				break
			}
			b.WriteString("  • " + k.What + ": " + k.Why + "\n")
		}
	}
	if checked != "" {
		say(&b, "adopt.done.open", "file", convert.ReportFile)
	} else {
		say(&b, "adopt.done.open.first")
	}
	return b.String()
}
