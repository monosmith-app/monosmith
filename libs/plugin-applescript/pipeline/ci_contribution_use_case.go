package pipeline

import (
	"os/exec"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CI contributes nothing: osacompile ships with macOS, and there is nothing
// to install on a macOS runner. The projects are tagged native, so a pipeline
// builds them in its native job; [ci] native_runners = ["macos-latest"] keeps
// that job to the one system that can.
func CI(protocol.CIParams) protocol.CIResult { return protocol.CIResult{} }

// lookPath finds a tool on PATH; tests replace it.
var lookPath = exec.LookPath

// Doctor warns when osacompile is not on PATH, which is any system but
// macOS: the projects are listed and versioned, but cannot be built there.
func Doctor(_ string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, err := lookPath("osacompile")
	return protocol.DoctorResult{Findings: []protocol.Finding{{
		Check: "osacompile is installed", OK: err == nil, Warning: true,
		Detail: "osacompile is not on PATH: AppleScript projects build on macOS only",
		Remedy: "build on a Mac; in CI set [ci] native_runners = [\"macos-latest\"] in smith.toml",
	}}}
}
