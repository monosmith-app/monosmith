package pipeline

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// SwiftVersion is the toolchain CI installs. Swift 6 ships swift format, which
// the lint task runs; bump it deliberately.
const SwiftVersion = "6.0"

// swift is installed by swift-actions/setup-swift on GitHub. Elsewhere
// swift.org publishes tarballs per distribution and patch release, which a
// generic step cannot name, so the step checks for Swift and says where to get it.
var swift = protocol.Toolchain{
	Name: "swift", Version: SwiftVersion,
	GitHub: &protocol.ActionStep{Uses: "swift-actions/setup-swift@v2", Inputs: []protocol.Input{{Name: "swift-version", Value: SwiftVersion}}},
	Install: "command -v swift >/dev/null || { echo 'Swift " + SwiftVersion + " is not installed: use an image that has it, or install it from https://www.swift.org/install/linux/' >&2; exit 1; }\n" +
		"swift --version",
}

// CI returns the Swift contribution: the Swift toolchain.
func CI(params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	return protocol.CIResult{Toolchains: []protocol.Toolchain{swift}}
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that Swift is installed.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, err := probe(workspaceRoot, []string{"swift", "--version"})
	return protocol.DoctorResult{Findings: []protocol.Finding{{
		Check: "Swift is installed", OK: err == nil,
		Detail: "swift --version failed",
		Remedy: "install Swift " + SwiftVersion + " or newer (https://www.swift.org/install/)",
	}}}
}
