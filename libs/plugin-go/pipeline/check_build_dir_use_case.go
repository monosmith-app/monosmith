package pipeline

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/taskoutput"
)

// goBuild is `go build` of one package. `go build ./...` compiles every
// package and writes no binary, so it is not a build into a folder.
var goBuild = taskoutput.Spec{
	Match: func(c []string) bool {
		return len(c) >= 2 && c[0] == "go" && c[1] == "build" && !slices.ContainsFunc(c[2:], func(a string) bool { return strings.Contains(a, "...") })
	},
	Flags:   []string{"-o"},
	Flag:    "-o",
	Default: "the binary into the folder it runs in",
	Value:   func(target string) string { return target + "/<binary>" },
}

// CheckBuildDirs reports a Go build task that writes outside the workspace's
// dist/. A task smith made passes -o {workspaceRoot}/dist/<project dir>/..., so
// only a build task written by hand in smith.toml is reported. The task's
// command is read, not the files of the module (`go build` writes only where
// -o says), and nothing is checked for a task that is not `go build`.
//
// Parameters: projects are the Go projects. Returns a finding for each
// project whose build task writes outside dist/; none when all is well.
func CheckBuildDirs(projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if f, ok := taskoutput.Check(p, "build", goBuild); ok {
			findings = append(findings, f)
		}
	}
	return findings
}
