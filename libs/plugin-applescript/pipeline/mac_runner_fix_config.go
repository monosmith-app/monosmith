package pipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// macRunnerFix is the fix of a project that builds on macOS only: set
// `[ci] native_runners` to the macOS runner with smith config set, then
// regenerate the pipelines so the native job keeps to that system. The host
// replaces the program "smith" with the smith that runs; the person reviews the
// plan of each step before it runs.
//
// Returns: the fix; it is never Safe, because it edits smith.toml and the pipelines.
func macRunnerFix() *protocol.Fix {
	return &protocol.Fix{
		Description: "Build the native projects on a macOS runner in CI",
		Steps: []protocol.FixStep{
			{Description: "set native_runners to macos-latest in smith.toml", Argv: []string{"smith", "config", "set", "ci.native_runners", "macos-latest"}},
			{Description: "regenerate the pipelines, so the native job runs on macOS", Argv: []string{"smith", "ci", "generate", "--yes"}},
		},
	}
}
