package adoptmove

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/adoptsplit"
)

// SortDependencies fills the plan's dependency lists from the manifest at the
// repository root, as the project's language declares them: the tools that
// go to the workspace, the ones its shared tooling replaces, and the ones
// nobody can place without asking. A language with no dependency split leaves
// the plan as it is. Nothing is changed.
//
// Parameters: root is the repository root; c the project; p the plan to fill.
// Returns: an error when the manifest cannot be read.
func SortDependencies(root string, c Candidate, p *Plan) error {
	sort := adoptsplit.For(c.Ecosystem).Sort
	if sort == nil {
		return nil
	}
	sorted, err := sort(root)
	if err != nil {
		return err
	}
	p.Hoist, p.Replaced, p.Unsure, p.Lint = sorted.Tools, sorted.Replaced, sorted.Unsure, sorted.Lint
	return nil
}

// Settle places each undecided dependency: the ones in workspace go to the
// workspace manifest, the rest stay in the project. It leaves no dependency
// undecided.
//
// Parameters: workspace are the names the person sent to the workspace.
func (p *Plan) Settle(workspace []string) {
	for _, name := range p.Unsure {
		if slices.Contains(workspace, name) {
			p.Hoist = append(p.Hoist, name)
		} else {
			p.Keep = append(p.Keep, name)
		}
	}
	slices.Sort(p.Hoist)
	p.Unsure = nil
}

// SettleLint records the answer about the project's own ESLint setup: kept
// in the project, or replaced by the shared config. Keeping it means no lint
// package is removed.
//
// Parameters: keep is true when the person keeps their own setup.
func (p *Plan) SettleLint(keep bool) {
	p.Lint = false
	if keep {
		p.KeepLint = true
		p.Replaced = nil
	}
}
