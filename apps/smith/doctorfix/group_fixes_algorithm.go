package doctorfix

import (
	"encoding/json"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Group turns the fixes of failed checks into the plan to offer:
//   - identical fixes carried by several checks are one fix
//   - fixes that name projects (AppliesTo) and have no Choices that share a Description are folded
//     into one that covers all of them (files and steps together, a step that
//     repeats runs once), unless two of them write the same file
//
// Items without a fix are skipped. The order is that of first appearance.
func Group(items []Item) []Planned {
	var plan []Planned
	byIdentity := map[string]int{}
	byKind := map[string]int{}
	for _, it := range items {
		if it.Fix == nil {
			continue
		}
		fix := *it.Fix
		foldable := len(fix.AppliesTo) > 0 && len(fix.Choices) == 0
		key := identity(fix)
		if i, ok := byIdentity[key]; ok {
			// The same change again (a fix for every project, carried by each
			// project's check): one fix, one more check.
			addCheck(&plan[i], it.Check)
			addProjects(&plan[i], fix.AppliesTo)
			continue
		}
		if foldable {
			if i, ok := byKind[fix.Description]; ok && disjoint(plan[i].Files, fix.Files) {
				fold(&plan[i], it.Check, fix)
				byIdentity[key] = i
				continue
			}
		}
		fix.Files, fix.Steps, fix.AppliesTo = slices.Clone(fix.Files), slices.Clone(fix.Steps), slices.Clone(fix.AppliesTo)
		p := Planned{Fix: fix, Checks: []string{it.Check}}
		if foldable {
			byKind[fix.Description] = len(plan)
		}
		byIdentity[key] = len(plan)
		p.Ask = Asks(p.Fix)
		plan = append(plan, p)
	}
	return plan
}

func fold(p *Planned, check string, fix protocol.Fix) {
	addCheck(p, check)
	addProjects(p, fix.AppliesTo)
	p.Files = append(p.Files, fix.Files...)
	for _, s := range fix.Steps {
		if !slices.ContainsFunc(p.Steps, func(o protocol.FixStep) bool { return sameStep(o, s) }) {
			p.Steps = append(p.Steps, s)
		}
	}
	p.Safe = p.Safe && fix.Safe
	p.Ask = Asks(p.Fix)
}

func addProjects(p *Planned, projects []string) {
	for _, project := range projects {
		if !slices.Contains(p.AppliesTo, project) {
			p.AppliesTo = append(p.AppliesTo, project)
		}
	}
}

func sameStep(a, b protocol.FixStep) bool {
	return a.Dir == b.Dir && slices.Equal(a.Argv, b.Argv)
}

func addCheck(p *Planned, check string) {
	if !slices.Contains(p.Checks, check) {
		p.Checks = append(p.Checks, check)
	}
}

func disjoint(a, b []protocol.FileEdit) bool {
	for _, x := range a {
		for _, y := range b {
			if x.Path == y.Path {
				return false
			}
		}
	}
	return true
}

// identity is what makes two fixes the same change.
func identity(f protocol.Fix) string {
	f.AppliesTo = nil
	raw, _ := json.Marshal(f)
	return string(raw)
}
