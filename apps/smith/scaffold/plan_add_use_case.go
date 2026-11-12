package scaffold

import (
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
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
	languageName := ""
	if known {
		if language, st, err := chooseStack(s, opts); err == nil {
			languageName = language.Info.Plugin.Name
			if language.Info.Plugin.Name == "node" {
				dir = nodeWorkspaceDirectory(s.Root, dir, opts.Name)
			}
			stackPlugins = st.Plugins()
			plan.Add("Project "+opts.Name, operationplan.Line{Action: operationplan.Create, Path: dir + "/" + opts.Name + "/", Note: "made by the " + language.Info.Plugin.Name + " plugin" + packageNameNote(opts.PackageName) + followNote(s, opts, kind, language.Info.Plugin.Name)})
			for _, name := range stackPlugins {
				plan.Add("Added by the "+name+" plugin", operationplan.Line{Action: operationplan.Change, Path: dir + "/" + opts.Name + "/", Note: "its files and tasks"})
			}
		} else {
			known = false
		}
	}
	if !known {
		note := "the stack decides the files: it comes from the recipe, the choices or the language" + packageNameNote(opts.PackageName)
		plan.Add("Project "+opts.Name, operationplan.Line{Action: operationplan.Create, Path: dir + "/" + opts.Name + "/", Note: note})
	}
	if opts.Environments != nil {
		for _, l := range EnvironmentLines(dir+"/"+opts.Name, *opts.Environments) {
			plan.Add("Environments: one build each, with its own .env file", l)
		}
	}
	if len(opts.With) > 0 || s.Config.Workspace.Recipe != "" && opts.Stack == "" {
		plan.Add("Workspace", operationplan.Line{Action: operationplan.Change, Path: "smith.toml", Note: "the stack of the project is saved"})
	}
	removesScript := false
	if known && cipipeline.PackageManagerInstallsSmith(languageName) {
		removesScript = slices.Contains(cipipeline.GeneratedFiles(s.Root), cipipeline.InstallScriptPath)
	}
	note := "the pipelines install smith with the package manager of this project, so the script is not used any more; the empty folder goes with it"
	if !known && slices.Contains(cipipeline.GeneratedFiles(s.Root), cipipeline.InstallScriptPath) {
		removesScript = true
		note = "removed only if the project uses node, python or .NET: the pipelines then install smith with its package manager; the empty folder goes with it"
	}
	if removesScript {
		plan.Add("Generated files smith removes", operationplan.Line{Action: operationplan.Remove, Path: cipipeline.InstallScriptPath, Note: note})
	}
	first := true
	for _, f := range cipipeline.GeneratedFiles(s.Root) {
		if removesScript && f == cipipeline.InstallScriptPath {
			continue
		}
		// The explanation is said once: a long note repeated for every file
		// pushes the commit box and the command out of the confirmation.
		note := "regenerated"
		if first {
			note = "regenerated after the add; edits outside smith:slot markers are lost (smith ci generate --dry-run shows the diff)"
			first = false
		}
		plan.Add("Generated files smith may rewrite", operationplan.Line{Action: operationplan.Change, Path: f, Note: note})
	}
	return operationplan.WithOwners(s.Root, plan)
}

// packageNameNote is what the plan says of a package name the person gave:
// the name the project is published or imported under; empty when none.
func packageNameNote(name string) string {
	if name == "" {
		return ""
	}
	return ", package name " + name
}
