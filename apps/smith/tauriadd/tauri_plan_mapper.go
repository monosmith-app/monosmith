package tauriadd

import (
	"path"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

// PlanFor lists what smith tauri add does to a project, so the person sees it
// before anything is written.
//
// Parameters: project is the app's name; root is the app's folder; stack is the
// stack it will have; srcTauriExists says the app has a src-tauri already, whose
// files are left as they are; identifier is the bundle identifier ("" when the
// existing tauri.conf.json keeps its own).
// Returns: the plan to present.
func PlanFor(project, root, stack string, srcTauriExists bool, identifier string) operationplan.Plan {
	plan := operationplan.Plan{Title: "make " + project + " a Tauri desktop app"}
	if !srcTauriExists {
		for _, f := range []struct{ file, note string }{
			{"Cargo.toml", "its own empty [workspace]"},
			{"tauri.conf.json", "bundle identifier " + identifier},
			{"build.rs", ""}, {"src/main.rs", ""}, {"src/lib.rs", ""},
			{"capabilities/default.json", ""}, {"icons/", "placeholder icons; smith tauri icon makes yours"},
		} {
			plan.Add("New: the Rust side", operationplan.Line{Action: operationplan.Create, Path: path.Join(root, "src-tauri", f.file), Note: f.note})
		}
	}
	plan.Add("Edited", operationplan.Line{Action: operationplan.Change, Path: "package.json", Note: "@tauri-apps/cli joins the root devDependencies"})
	plan.Add("Edited", operationplan.Line{Action: operationplan.Change, Path: path.Join(root, "package.json"), Note: "@tauri-apps/api and a tauri script"})
	plan.Add("Edited", operationplan.Line{Action: operationplan.Change, Path: "smith.toml", Note: "stack " + stack + " for " + project})
	plan.Add("Then", operationplan.Line{Action: operationplan.Change, Path: "the installed packages", Note: "smith runs the install (network)"})
	plan.Add("Then", operationplan.Line{Action: operationplan.Change, Path: "the CI pipelines", Note: "Rust, and on Linux the system libraries; smith ci generate --dry-run shows the diff"})
	return plan
}
