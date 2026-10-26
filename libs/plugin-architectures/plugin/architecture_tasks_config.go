package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/imports"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the task every architecture adds to each project.
const TaskName = "architecture-check"

// architectureCheck runs smith's own checker on the project root. Only the
// sources' imports matter, so the cached result is reused exactly when a source
// file changed.
func architectureCheck(rs rules.Ruleset, lang imports.Language) protocol.Task {
	return protocol.Task{
		Command:     []string{"{smith}", "slices", "check", "--architecture", rs.Name, "--lang", lang.Name, "."},
		Inputs:      lang.Inputs,
		Cache:       true,
		Description: "checks the " + rs.Name + " architecture: " + rs.Summary,
	}
}
