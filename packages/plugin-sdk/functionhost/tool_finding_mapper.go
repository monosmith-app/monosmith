package functionhost

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// FindingFor is the doctor finding for a host's local tool. It is a warning,
// never a failure: the tests and the build need no cloud tool, only running the
// function locally does.
//
// Parameters: tool is the host's tool; installed says whether it is found.
// Returns: the finding, with the way to get the tool as its remedy.
func FindingFor(tool Tool, installed bool) protocol.Finding {
	f := protocol.Finding{
		Check:   tool.Name + " is installed",
		OK:      installed,
		Warning: true,
		Area:    "toolchain",
		Detail:  "the start task runs the function locally with " + tool.Name + ", and it is not found",
		Remedy:  tool.Install,
	}
	if !installed {
		f.Problem = tool.Name + " is not found, so the function cannot run locally"
	}
	return f
}
