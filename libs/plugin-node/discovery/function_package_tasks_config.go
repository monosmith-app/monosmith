package discovery

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PackageCheckTask is the task that checks a function app's package without
// installing its dependencies or writing its zip: smith verify runs it in the
// place of the package task.
const PackageCheckTask = "package-check"

// PrebuiltEnv tells smith func package that the task graph built the app and
// its libraries first, so it does not build them again.
const PrebuiltEnv = "SMITH_FUNC_PREBUILT"

// FunctionPackage says whether a package script packs a function app with
// smith func package.
func FunctionPackage(script string) bool {
	return strings.HasPrefix(strings.TrimSpace(script), "smith func package")
}

// ApplyFunctionPackage makes the package task of a function app cacheable
// and adds its check:
//
//   - package runs after the app's build and its libraries' builds, does not
//     build them again (PrebuiltEnv), and caches the build folder with the zip
//     and the copy smith ci pack keeps in dist/drop/<app>, so a run whose
//     inputs did not change restores them instead of installing again;
//   - package-check lays out the stage and checks that every import is
//     declared, without installing or zipping (smith func package --check).
//     It runs in smith verify, where package used to.
//
// Parameters: tasks are the project's tasks, package among them; root is the
// workspace-relative project folder; scripts are its package.json scripts.
func ApplyFunctionPackage(tasks map[string]protocol.Task, root string, scripts map[string]string) {
	deps := []string{"^build"}
	if _, builds := scripts["build"]; builds {
		deps = []string{"build", "^build"}
	}
	env := map[string]string{PrebuiltEnv: "1"}
	pack := tasks["package"]
	pack.Cache = true
	pack.DependsOn = deps
	pack.Inputs = workspaceInputs
	pack.Env = env
	pack.Outputs = []string{protocol.BuildDirFromProject(root), protocol.FromProject(root, "dist/drop/"+path.Base(strings.Trim(root, "/")))}
	pack.Description = "Pack the function app into its zip and dist/drop (smith ci pack), after the builds it needs"
	tasks["package"] = pack
	tasks[PackageCheckTask] = protocol.Task{
		Command:     []string{"{smith}", "func", "package", "--check"},
		DependsOn:   deps,
		Cache:       true,
		Inputs:      workspaceInputs,
		Env:         env,
		Verify:      true,
		Description: "Check the function app's package without installing or zipping: the stage, the compiled entry, and the dependencies it and its libraries declare (smith verify)",
	}
}
