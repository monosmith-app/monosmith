package depsrepair

import (
	"fmt"
	"slices"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// askTimeout bounds one dependencies/repair call: the plugin only reads the disk.
const askTimeout = time.Minute

// asker is one plugin that has the repair capability, with the projects the
// repair is for. gather reads the plans through it, so a test needs no plugin.
type asker struct {
	plugin   string
	projects []protocol.Project
	// installs is true when the plugin can install (dependencies/install).
	installs bool
	ask      func(protocol.RepairParams) (protocol.RepairResult, error)
}

// Load asks every plugin of the workspace that has the repair capability how
// it repairs the dependencies of the selected projects.
//
// Parameters: s is the open workspace; targets select projects the way
// `smith run` does (names, globs, lang:<language>); none selects every project.
//
// Returns: the offer. It has no tool when no selected project belongs to a
// plugin that repairs; Offer.Notes then says which plugins do not.
//
// Errors: a target that matches no project.
func Load(s *session.Session, targets []string) (Offer, error) {
	names, err := s.Graph.Select(targets)
	if err != nil {
		return Offer{}, err
	}
	var roots []string
	for _, p := range s.Graph.Projects {
		roots = append(roots, p.Root)
	}
	var askers []asker
	var without []string
	for _, c := range s.Plugins.Clients {
		plugin := c.Info.Plugin.Name
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.Plugin == plugin && slices.Contains(names, p.Name) {
				projects = append(projects, p.Project)
			}
		}
		if len(projects) == 0 {
			continue
		}
		if !c.Supports(protocol.MethodRepair) {
			if len(targets) > 0 && c.Supports(protocol.MethodInstall) {
				// The person named these projects: say why nothing is offered for them.
				without = append(without, plugin)
			}
			continue
		}
		client := c
		askers = append(askers, asker{
			plugin: plugin, projects: projects, installs: c.Supports(protocol.MethodInstall),
			ask: func(params protocol.RepairParams) (protocol.RepairResult, error) {
				var res protocol.RepairResult
				err := client.Call(protocol.MethodRepair, params, &res, askTimeout)
				return res, err
			},
		})
	}
	offer := gather(s.Root, roots, askers)
	for _, plugin := range without {
		offer.Notes = append(offer.Notes, fmt.Sprintf("%s has no dependency repair: its plugin does not say how its tool repairs", plugin))
	}
	return offer, nil
}

// gather reads the plans of each asker and keeps the ones smith accepts. A plan
// is refused, with a note that says why, when its kind is not one of the
// protocol, when the plugin gave the kind twice, when it does nothing, when it
// asks for an install the plugin cannot do, or when it names a path smith does
// not delete (see checkDeletePath). A plugin that fails to answer gives a note
// and no tool, so the other tools are still offered.
func gather(root string, projectRoots []string, askers []asker) Offer {
	offer := Offer{Root: root, ProjectRoots: projectRoots}
	order := protocol.RepairKinds()
	for _, a := range askers {
		res, err := a.ask(protocol.RepairParams{Projects: a.projects})
		if err != nil {
			offer.Notes = append(offer.Notes, fmt.Sprintf("%s could not say how it repairs dependencies: %v", a.plugin, err))
			continue
		}
		tool := Tool{Plugin: a.plugin, Projects: a.projects}
		for _, kind := range order {
			seen := false
			for _, plan := range res.Plans {
				if plan.Kind != kind {
					continue
				}
				if seen {
					offer.Notes = append(offer.Notes, fmt.Sprintf("%s gave two plans for %s: smith uses the first", a.plugin, kind))
					continue
				}
				seen = true
				if why := refusal(root, projectRoots, a, plan); why != "" {
					offer.Notes = append(offer.Notes, fmt.Sprintf("%s: %s is not offered: %s", a.plugin, kind, why))
					continue
				}
				tool.Plans = append(tool.Plans, plan)
			}
		}
		for _, plan := range res.Plans {
			if !slices.Contains(order, plan.Kind) {
				offer.Notes = append(offer.Notes, fmt.Sprintf("%s gave a plan of the kind %q, which smith does not know", a.plugin, plan.Kind))
			}
		}
		if len(tool.Plans) > 0 {
			offer.Tools = append(offer.Tools, tool)
		}
	}
	return offer
}

// refusal says why smith does not offer a plan ("" when it does).
func refusal(root string, projectRoots []string, a asker, plan protocol.RepairPlan) string {
	if len(plan.Delete) == 0 && len(plan.Commands) == 0 && !plan.Install {
		return "the plan deletes nothing, runs nothing and installs nothing"
	}
	if plan.Install && !a.installs {
		return "the plan installs again, and the plugin cannot install"
	}
	for _, rel := range plan.Delete {
		if _, err := checkDeletePath(root, rel, projectRoots); err != nil {
			return err.Error()
		}
	}
	for _, cmd := range plan.Commands {
		if len(cmd.Argv) == 0 || cmd.Argv[0] == "" {
			return "a command of the plan has no program"
		}
		if cmd.Dir != "" && !insideWorkspace(cmd.Dir) {
			return fmt.Sprintf("a command runs in %q, which is not a folder of the workspace", cmd.Dir)
		}
	}
	return ""
}
