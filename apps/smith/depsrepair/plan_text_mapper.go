package depsrepair

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// bannerWidth is the width of the banner rule, as the other commands draw it.
const bannerWidth = 60

// KindLabel is the name of a repair kind as a person reads it ("Clean reinstall").
//
// Parameters: kind is the repair kind.
func KindLabel(kind protocol.RepairKind) string {
	return words.Say("depsrepair.kind." + string(kind) + ".label")
}

// KindExplanation says in one sentence what a repair kind deletes and runs,
// and what it does to the versions.
//
// Parameters: kind is the repair kind.
func KindExplanation(kind protocol.RepairKind) string {
	return words.Say("depsrepair.kind." + string(kind) + ".explain")
}

// NoToolText is the reason an interface gives for a kind no tool has.
func NoToolText() string { return words.Say("depsrepair.kind.none") }

// Banner is the first line of the command: "══ Dependency repair: Prune ═══".
//
// Parameters: kind is the repair kind, or "" for the banner of every kind (a
// dry run without --kind).
func Banner(kind protocol.RepairKind) string {
	title := "Dependency repair"
	if kind != "" {
		title += ": " + KindLabel(kind)
	}
	rule := "══ " + title + " "
	return rule + strings.Repeat("═", max(bannerWidth-len([]rune(rule)), 3))
}

// ChoiceLines are the numbered lines a terminal shows to pick a kind: every
// kind of the protocol, lightest first, with what it does. A kind no tool has
// says so.
//
// Parameters: offer is the workspace's offer.
func ChoiceLines(offer Offer) []string {
	var out []string
	for i, kind := range protocol.RepairKinds() {
		line := fmt.Sprintf("  %d) %s: %s", i+1, KindLabel(kind), KindExplanation(kind))
		if !offer.Has(kind) {
			line += " (" + NoToolText() + ")"
		}
		out = append(out, line)
	}
	return out
}

// PlanText is what a repair kind deletes, runs and installs, tool by tool,
// then the warnings, each with its remedy: the text a terminal prints before it
// asks and the TUI shows as the review.
//
// Parameters: offer is the workspace's offer; kind is the repair kind.
func PlanText(offer Offer, kind protocol.RepairKind) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s\n", KindLabel(kind), KindExplanation(kind))
	for _, t := range offer.Tools {
		plan, ok := t.Plan(kind)
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "\n%s (%s): %s\n", t.Plugin, projectCount(len(t.Projects)), plan.Summary)
		for _, p := range plan.Delete {
			fmt.Fprintf(&b, "  delete   %s\n", p)
		}
		for _, c := range plan.Commands {
			line := "  run      " + strings.Join(c.Argv, " ")
			if c.Dir != "" {
				line += "   (in " + c.Dir + ")"
			}
			b.WriteString(line + "\n")
		}
		if plan.Install {
			fmt.Fprintf(&b, "  install  %s, as smith install does\n", projectCount(len(t.Projects)))
		}
		for _, n := range plan.Notes {
			fmt.Fprintf(&b, "  ● %s\n", n)
		}
	}
	var others []string
	for _, t := range offer.Tools {
		if _, ok := t.Plan(kind); !ok {
			others = append(others, t.Plugin)
		}
	}
	if len(others) > 0 {
		fmt.Fprintf(&b, "\n● %s: no such repair, left as it is\n", strings.Join(others, ", "))
	}
	warnings := false
	warn := func(key string) {
		if !warnings {
			b.WriteString("\n")
			warnings = true
		}
		fmt.Fprintf(&b, "%s %s\n  → %s\n", marks.Warning, words.Say(key), words.Say(key+".remedy"))
	}
	if offer.MachineWide(kind) {
		warn("depsrepair.warn.machine_wide")
	}
	if offer.ChangesVersions(kind) {
		warn("depsrepair.warn.versions")
	}
	for _, n := range offer.Notes {
		fmt.Fprintf(&b, "\n● %s", n)
	}
	if len(offer.Notes) > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

// Closing is the last line of a repair that ended well.
//
// Parameters: offer is the workspace's offer; kind is the repair kind.
func Closing(offer Offer, kind protocol.RepairKind) string {
	var tools []string
	for _, t := range offer.Tools {
		if _, ok := t.Plan(kind); ok {
			tools = append(tools, t.Plugin)
		}
	}
	return fmt.Sprintf("%s Repaired the dependencies (%s) with %s. Nothing is committed.", marks.Passed, strings.ToLower(KindLabel(kind)), strings.Join(tools, ", "))
}

// projectCount is "1 project" or "3 projects".
func projectCount(n int) string {
	if n == 1 {
		return "1 project"
	}
	return fmt.Sprintf("%d projects", n)
}
