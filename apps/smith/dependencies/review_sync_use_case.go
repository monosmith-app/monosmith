package dependencies

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ErrNotConfirmed is the person answering no to the sync plan.
var ErrNotConfirmed = errors.New("cancelled: nothing was changed")

// Reviewer settles a sync plan before anything is written.
type Reviewer struct {
	In  *bufio.Reader
	Out io.Writer
	// Yes is --yes: write the plan as it stands.
	Yes bool
	// Asks is true when a person can answer (a terminal, not CI).
	Asks bool
	// Install says the plan ends with an install.
	Install bool
	// Preview is --dry-run: print the plan and ask nothing.
	Preview bool
	// CrossMajor is --major: a declaration may move to another major version.
	CrossMajor bool
	// Prefer is --prefer: highest, lowest or installed for every dependency.
	Prefer string
	// To is --to name=spec: the version of single dependencies.
	To map[string]string
}

// Settle applies the flags to each dependency, prints the plan, and in a
// terminal asks which version wins and whether to write. With --yes it prints
// the plan and goes on. Without a terminal and without --yes it refuses.
//
// Parameters: drifts are the dependencies the projects disagree on, with
// smith's pick as target.
// Returns: the drifts with their final targets; ErrNotConfirmed when the person
// says no; a refusal that names the flags when nobody can answer.
func (r Reviewer) Settle(drifts []Drift) ([]Drift, error) {
	open := map[int]bool{}
	for i, d := range drifts {
		if spec, ok := r.To[d.Name]; ok {
			drifts[i] = Retarget(d, spec)
			continue
		}
		if r.Prefer != PreferPick {
			if next, ok := Prefer(d, r.Prefer); ok {
				drifts[i] = next
				continue
			}
			fmt.Fprintf(r.Out, "note: %s has no %s version, so it keeps smith's pick (%s)\n", d.Name, r.Prefer, d.Target)
		}
		open[i] = true
	}
	r.normalise(drifts)
	RenderSyncPlan(r.Out, drifts, r.Install)
	if r.Preview {
		fmt.Fprintln(r.Out, "\nDry run: nothing was changed.")
		return drifts, nil
	}
	if r.Yes {
		fmt.Fprintln(r.Out, "--yes: this plan is written as it stands. --prefer highest|lowest|installed or --to <name>=<version> changes a version.")
		return drifts, nil
	}
	if !r.Asks {
		return drifts, problem.New("nobody can confirm this change: there is no terminal and --yes was not given",
			"--yes writes the plan above",
			"--prefer highest|lowest|installed picks the winner for every dependency",
			"--to <name>=<version> picks it for one",
			"--dry-run shows the plan and changes nothing")
	}
	if len(open) > 0 && slices.ContainsFunc(drifts, func(d Drift) bool { return len(Candidates(d)) > 1 }) {
		answer := choiceprompt.Ask("These dependencies are declared at different versions. Which version wins?", []choiceprompt.Option{
			{Label: "smith's pick for each", Value: "pick", Hint: "the installed version, else the highest declared (default)"},
			{Label: "The highest declared", Value: PreferHighest, Hint: "projects on an older version move up"},
			{Label: "The lowest declared", Value: PreferLowest, Hint: "projects on a newer version move down"},
			{Label: "Choose one by one", Value: "each", Hint: "smith asks for each dependency"},
			{Label: "Cancel", Value: "cancel", Hint: "nothing is changed"},
		}, "pick", r.In, r.Out)
		switch answer {
		case "cancel":
			return drifts, ErrNotConfirmed
		case PreferHighest, PreferLowest:
			for i := range open {
				if next, ok := Prefer(drifts[i], answer); ok {
					drifts[i] = next
				}
			}
		case "each":
			for i := range drifts {
				if !open[i] {
					continue
				}
				drifts[i] = r.askOne(drifts[i])
			}
		}
		fmt.Fprintln(r.Out, "\nThe plan with your choices:")
		RenderSyncPlan(r.Out, drifts, r.Install)
	}
	writes := 0
	files := map[string]bool{}
	for _, d := range drifts {
		for _, s := range d.Fixable {
			writes++
			files[s.File] = true
		}
	}
	tail := ""
	if r.Install {
		tail = " and run the install (network; it can take minutes)"
	}
	fmt.Fprintf(r.Out, "\nWrite %d declaration(s) in %d file(s)%s? [y/N] ", writes, len(files), tail)
	answer, _ := r.In.ReadString('\n')
	if !strings.EqualFold(strings.TrimSpace(answer), "y") {
		return drifts, ErrNotConfirmed
	}
	return drifts, nil
}

// askOne asks which version wins for one dependency.
func (r Reviewer) askOne(d Drift) Drift {
	cands := Candidates(d)
	if len(cands) < 2 {
		return d
	}
	options := make([]choiceprompt.Option, len(cands))
	for i, c := range cands {
		options[i] = choiceprompt.Option{Label: c.Spec, Value: c.Spec, Hint: c.Why}
	}
	spec := choiceprompt.Ask("Which version of "+d.Name+" wins?", options, cands[0].Spec, r.In, r.Out)
	if spec == cands[0].Spec {
		return d
	}
	return Retarget(d, spec)
}

// RenderSyncPlan prints, for each dependency, what every project declares now
// and what it will declare, so the person sees the change before it happens.
//
// Parameters: w is the output; drifts are the dependencies; install says the
// plan ends with an install.
func RenderSyncPlan(w io.Writer, drifts []Drift, install bool) {
	if len(drifts) == 0 {
		fmt.Fprintln(w, "Every project agrees on every shared dependency version. Nothing to change.")
		return
	}
	fmt.Fprintf(w, "%d dependencies are declared at different versions:\n", len(drifts))
	for _, d := range drifts {
		fmt.Fprintf(w, "\n%s (%s) will be %s  %s\n", d.Name, d.Plugin, d.Target, hint(d.Reason))
		for _, s := range d.Sites {
			switch {
			case s.Spec == d.Target:
				fmt.Fprintf(w, "  %-12s %s (stays)\n", where(s), s.Spec)
			case slices.ContainsFunc(d.Fixable, func(f protocol.Declaration) bool { return same(f, s) }):
				fmt.Fprintf(w, "  %-12s %s -> %s (%s)\n", where(s), s.Spec, d.Target, s.File)
			case slices.ContainsFunc(d.Crossing, func(f protocol.Declaration) bool { return same(f, s) }):
				fmt.Fprintf(w, "  %-12s %s stays: moving it from major %d to %d can break the project (--major allows it)\n", where(s), s.Spec, majorOf(s.Spec), majorOf(d.Target))
			default:
				fmt.Fprintf(w, "  %-12s %s stays: smith does not rewrite this kind of spec\n", where(s), s.Spec)
			}
		}
	}
	if install {
		fmt.Fprintln(w, "\nThen smith runs the install, so lockfiles and the venv follow.")
	}
}

// same reports whether two declarations are the same line of the same manifest.
func same(a, b protocol.Declaration) bool {
	return a.Project == b.Project && a.File == b.File && a.Name == b.Name && a.Section == b.Section && a.Spec == b.Spec
}

// normalise lets the declarations that cross a major version be written when
// --major is given, so the plan shows what will really happen.
func (r Reviewer) normalise(drifts []Drift) {
	if !r.CrossMajor {
		return
	}
	for i := range drifts {
		drifts[i].Fixable = append(drifts[i].Fixable, drifts[i].Crossing...)
		drifts[i].Crossing = nil
	}
}
