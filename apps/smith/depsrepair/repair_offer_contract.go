package depsrepair

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Offer is what the tools of a workspace can do to repair its installed
// dependencies: for each tool, one plan for each repair kind it has.
type Offer struct {
	// Root is the absolute workspace root every path of a plan is relative to.
	Root string
	// Tools are the plugins that offer a repair for their projects, in the
	// order the workspace loads them.
	Tools []Tool
	// ProjectRoots are the roots of every project of the workspace,
	// workspace-relative. A repair never deletes one of them, or a folder that
	// holds one.
	ProjectRoots []string
	// Notes say what was left out and why (a plan smith refused, a plugin that
	// could not answer). An interface shows them with the plan.
	Notes []string
}

// Tool is one plugin's part of an offer.
type Tool struct {
	// Plugin is the plugin's name ("node").
	Plugin string
	// Projects are the plugin's projects the repair is for.
	Projects []protocol.Project
	// Plans are the plans smith accepted, lightest kind first, one for a kind.
	Plans []protocol.RepairPlan
}

// Plan is the tool's plan for a kind.
//
// Parameters: kind is the repair kind.
//
// Returns: the plan, and false when the tool does not have the kind.
func (t Tool) Plan(kind protocol.RepairKind) (protocol.RepairPlan, bool) {
	for _, p := range t.Plans {
		if p.Kind == kind {
			return p, true
		}
	}
	return protocol.RepairPlan{}, false
}

// Kinds lists the repair kinds at least one tool has, lightest first.
//
// Returns: the kinds, empty when no tool of the workspace offers a repair.
func (o Offer) Kinds() []protocol.RepairKind {
	var out []protocol.RepairKind
	for _, kind := range protocol.RepairKinds() {
		if o.Has(kind) {
			out = append(out, kind)
		}
	}
	return out
}

// Has reports whether at least one tool has the kind.
//
// Parameters: kind is the repair kind.
func (o Offer) Has(kind protocol.RepairKind) bool {
	for _, t := range o.Tools {
		if _, ok := t.Plan(kind); ok {
			return true
		}
	}
	return false
}

// MachineWide reports whether the kind changes something outside the workspace
// with at least one tool (a download cache every workspace of the machine uses).
//
// Parameters: kind is the repair kind.
func (o Offer) MachineWide(kind protocol.RepairKind) bool {
	for _, t := range o.Tools {
		if p, ok := t.Plan(kind); ok && p.MachineWide {
			return true
		}
	}
	return false
}

// ChangesVersions reports whether the kind can install other versions than the
// ones installed now with at least one tool.
//
// Parameters: kind is the repair kind.
func (o Offer) ChangesVersions(kind protocol.RepairKind) bool {
	for _, t := range o.Tools {
		if p, ok := t.Plan(kind); ok && p.ChangesVersions {
			return true
		}
	}
	return false
}
