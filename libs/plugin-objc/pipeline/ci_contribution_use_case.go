package pipeline

import (
	"os/exec"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// cocoapods is installed when `pod` is missing (GitHub's macOS images carry
// it). Projects are tagged native, so it is the native job on a macOS runner
// that runs it.
var cocoapods = protocol.Toolchain{
	Name: "cocoapods", Version: "latest",
	Install: "command -v pod >/dev/null || gem install cocoapods --no-document",
}

// CI returns the Objective-C contribution: CocoaPods. The projects are
// tagged native, so the pipelines build them in their native job, which
// [ci] native_runners = ["macos-latest"] keeps to the system that has Xcode.
func CI(params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	return protocol.CIResult{Toolchains: []protocol.Toolchain{cocoapods}}
}

// lookPath finds a tool on PATH; tests replace it.
var lookPath = exec.LookPath

// Doctor warns when CocoaPods or Xcode's xcodebuild is not on PATH: the
// projects are listed and versioned everywhere, but linted and built on a Mac.
func Doctor(_ string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, podErr := lookPath("pod")
	_, xcodeErr := lookPath("xcodebuild")
	return protocol.DoctorResult{Findings: []protocol.Finding{
		{Check: "CocoaPods is installed", OK: podErr == nil, Warning: true, Detail: "pod is not on PATH", Remedy: "run gem install cocoapods (or brew install cocoapods) on a Mac"},
		{Check: "Xcode is installed", OK: xcodeErr == nil, Warning: true, Detail: "xcodebuild is not on PATH: Objective-C projects build on macOS only", Remedy: "build on a Mac with Xcode; in CI set [ci] native_runners = [\"macos-latest\"] in smith.toml"},
	}}
}
