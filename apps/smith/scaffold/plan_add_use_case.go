package scaffold

import (
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PlanAdd is what smith add will write, before it writes: the project
// directory, the part each plugin of the stack adds (lint, tests, build), and
// the workspace files that change. It asks nothing and writes nothing. When the
// stack depends on an answer that is still to come (the workspace recipe, --with
// choices, a language chosen from several), the plan lists the directory and the
// workspace files only, and says so: the plugins appear once the stack is known.
//
// Parameters: s is the open session; opts are the options smith add runs with.
// Returns: the plan, with owners filled from smith.lock; empty for a kind that
// AddProject will refuse.
func PlanAdd(s *session.Session, opts AddOptions) operationplan.Plan {
	kind := protocol.ProjectKind(opts.Kind)
	switch kind {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
	default:
		return operationplan.Plan{}
	}
	plan := operationplan.Plan{Title: "add " + opts.Kind + " " + opts.Name}
	dir := projectkind.Directory(kind)
	known := len(opts.With) == 0 && (opts.Stack != "" || opts.Lang != "") && (opts.Stack != "" || s.Config.Workspace.Recipe == "")
	var stackPlugins []string
	if known {
		if language, st, err := chooseStack(s, opts); err == nil {
			if language.Info.Plugin.Name == "node" {
				dir = nodeWorkspaceDirectory(s.Root, dir, opts.Name)
			}
			stackPlugins = st.Plugins()
			plan.Add("Project "+opts.Name, operationplan.Line{Action: operationplan.Create, Path: dir + "/" + opts.Name + "/", Note: "made by the " + language.Info.Plugin.Name + " plugin"})
			for _, name := range stackPlugins {
				plan.Add("Added by the "+name+" plugin", operationplan.Line{Action: operationplan.Change, Path: dir + "/" + opts.Name + "/", Note: "its files and tasks"})
			}
		} else {
			known = false
		}
	}
	if !known {
		note := "the stack decides the files: it comes from the recipe, the choices or the language"
		plan.Add("Project "+opts.Name, operationplan.Line{Action: operationplan.Create, Path: dir + "/" + opts.Name + "/", Note: note})
	}
	if len(opts.With) > 0 || s.Config.Workspace.Recipe != "" && opts.Stack == "" {
		plan.Add("Workspace", operationplan.Line{Action: operationplan.Change, Path: "smith.toml", Note: "the stack of the project is saved"})
	}
	plan.Add("Workspace", operationplan.Line{Action: operationplan.Change, Path: "CI pipelines", Note: "regenerated to install the new toolchain"})
	return operationplan.WithOwners(s.Root, plan)
}
