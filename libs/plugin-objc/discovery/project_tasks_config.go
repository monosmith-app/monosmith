package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// NativeTag marks the project as built on one operating system only: pod lib
// lint builds with Xcode, which runs on macOS alone, so the generated pipelines
// leave the project to their native job, which [ci] native_runners can
// restrict to macos-latest.
const NativeTag = "native"

// inputs are what a project's results depend on: its podspec and its sources.
var inputs = []string{"*.podspec", "**/*.h", "**/*.m", "**/*.mm", "**/*.c"}

// Tasks are CocoaPods' for one project, run in it: lint checks the podspec
// without downloading or building anything (--quick), and build is the full
// pod lib lint, which builds the pod with xcodebuild. Neither writes into the
// workspace: CocoaPods works in a temporary directory.
func Tasks(podspec string) map[string]protocol.Task {
	return map[string]protocol.Task{
		"lint":  {Command: []string{"pod", "lib", "lint", "--quick", podspec}, Inputs: inputs, Cache: true, Description: "pod lib lint --quick: check the podspec"},
		"build": {Command: []string{"pod", "lib", "lint", podspec}, Inputs: inputs, Cache: true, Description: "pod lib lint: build the pod with Xcode (macOS only)"},
	}
}
