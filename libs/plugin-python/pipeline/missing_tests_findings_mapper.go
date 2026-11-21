package pipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// MissingTests returns a warning for each project that has no test task.
// Discovery gives no test task to a project with no test file, because pytest
// would collect nothing and fail; the warning says so in place of a failed run.
//
// Parameters: projects are the Python projects of the workspace.
// Returns: one warning finding per project without a test task.
func MissingTests(projects []protocol.Project) []protocol.Finding {
	var out []protocol.Finding
	for _, p := range projects {
		if _, ok := p.Tasks["test"]; ok {
			continue
		}
		out = append(out, protocol.Finding{
			Check:   "the project has tests",
			Project: p.Name,
			Warning: true,
			Problem: "has no tests: no test_*.py or *_test.py file, so it has no test task",
			Detail:  p.Name + " has no test_*.py or *_test.py file, and pytest finds no tests to run",
			Remedy:  "add a test, such as tests/test_" + p.Name + ".py; smith run test -p " + p.Name + " then runs it with pytest",
		})
	}
	return out
}
