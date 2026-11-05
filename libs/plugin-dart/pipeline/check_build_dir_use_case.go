package pipeline

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/taskoutput"
)

// dartBuild is the build commands that take an output option: dart compile,
// dart build and flutter build.
var dartBuild = taskoutput.Spec{
	Match: func(c []string) bool {
		return len(c) >= 2 && (c[0] == "dart" && (c[1] == "compile" || c[1] == "build") || c[0] == "flutter" && c[1] == "build")
	},
	Flags:   []string{"-o", "--output"},
	Flag:    "-o",
	Default: "beside the entry file, or into build/",
	Value:   func(target string) string { return target },
}

// CheckBuildDirs reports a Dart or Flutter build task that writes outside the
// workspace's dist/. A task smith made passes -o {workspaceRoot}/dist/<project
// dir>, so only a build task written by hand in smith.toml is reported. The
// build/ folder that `flutter run` and the IDE use is not read.
//
// Parameters: projects are the Dart projects. Returns a finding for each
// project whose build task writes outside dist/; none when all is well.
func CheckBuildDirs(projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if f, ok := taskoutput.Check(p, "build", dartBuild); ok {
			findings = append(findings, f)
		}
	}
	return findings
}
