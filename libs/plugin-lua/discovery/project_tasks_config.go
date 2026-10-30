package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// inputs are what a project's results depend on: its Lua sources, its
// rockspec and the luacheck rules.
var inputs = []string{"**/*.lua", "*.rockspec", ".luacheckrc"}

// Tasks are the Lua tools' for one project, run in it: luacheck lints, busted
// runs spec/ when the project has one, and luarocks make installs the rock
// into dist/<project dir> (smith empties a task's outputs before it runs).
func Tasks(root, rockspec string, spec bool) map[string]protocol.Task {
	out := protocol.BuildDirFromProject(root)
	tasks := map[string]protocol.Task{
		"lint":  {Command: []string{"luacheck", "."}, Inputs: inputs, Cache: true, Description: "luacheck (.luacheckrc's rules)"},
		"build": {Command: []string{"luarocks", "make", "--tree", out, rockspec}, Inputs: inputs, Outputs: []string{out}, Cache: true, Description: "luarocks make: install the rock into dist/"},
	}
	if spec {
		tasks["test"] = protocol.Task{Command: []string{"busted"}, Inputs: inputs, Cache: true, Description: "busted (spec/)"}
	}
	return tasks
}
