package depsrepair

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Report is what `smith deps repair --json` prints: the plan of each kind
// asked for, and for a repair that ran, the kind it ran.
type Report struct {
	// Root is the workspace root.
	Root string `json:"root"`
	// Ran is the kind that was carried out; empty for a dry run.
	Ran protocol.RepairKind `json:"ran,omitempty"`
	// Kinds are the plans, one entry for each kind shown.
	Kinds []KindReport `json:"kinds"`
	// Notes say what was left out and why.
	Notes []string `json:"notes,omitempty"`
}

// KindReport is one repair kind across the tools of the workspace.
type KindReport struct {
	// Kind is the repair kind.
	Kind protocol.RepairKind `json:"kind"`
	// Label is the kind as a person reads it.
	Label string `json:"label"`
	// MachineWide is true when a tool's plan changes a cache other workspaces use.
	MachineWide bool `json:"machineWide,omitempty"`
	// ChangesVersions is true when a tool's plan can install other versions.
	ChangesVersions bool `json:"changesVersions,omitempty"`
	// Tools are the plans of the tools that have the kind.
	Tools []ToolReport `json:"tools"`
}

// ToolReport is one tool's plan for a kind.
type ToolReport struct {
	// Plugin is the plugin's name.
	Plugin string `json:"plugin"`
	// Projects are the names of the projects the repair is for.
	Projects []string `json:"projects"`
	// Plan is what the tool deletes, runs and installs.
	Plan protocol.RepairPlan `json:"plan"`
}

// reportOf is the report of the kinds, in the order given.
func reportOf(offer Offer, kinds []protocol.RepairKind, ran protocol.RepairKind) Report {
	r := Report{Root: offer.Root, Ran: ran, Notes: offer.Notes, Kinds: []KindReport{}}
	for _, kind := range kinds {
		k := KindReport{Kind: kind, Label: KindLabel(kind), MachineWide: offer.MachineWide(kind), ChangesVersions: offer.ChangesVersions(kind), Tools: []ToolReport{}}
		for _, t := range offer.Tools {
			plan, ok := t.Plan(kind)
			if !ok {
				continue
			}
			names := make([]string, len(t.Projects))
			for i, p := range t.Projects {
				names[i] = p.Name
			}
			k.Tools = append(k.Tools, ToolReport{Plugin: t.Plugin, Projects: names, Plan: plan})
		}
		r.Kinds = append(r.Kinds, k)
	}
	return r
}
