package pipeline

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/taskoutput"
)

// pythonBuild is `python -m build`, which writes the wheel and the sdist.
var pythonBuild = taskoutput.Spec{
	Match: func(c []string) bool {
		i := slices.Index(c, "-m")
		return i >= 0 && i+1 < len(c) && c[i+1] == "build"
	},
	Flags:   []string{"--outdir", "-o"},
	Flag:    "--outdir",
	Default: "dist/ inside the project",
	Value:   func(target string) string { return target },
}

// CheckBuildDirs reports a Python build task that writes outside the
// workspace's dist/. A task smith made passes --outdir {workspaceRoot}/dist/
// <project dir>, so only a build task written by hand in smith.toml is
// reported. The folders a build backend makes inside the project (setuptools'
// build/ and *.egg-info) are not read: `python -m build` does not say where it
// puts them, and no setting of smith moves them.
//
// Parameters: projects are the Python projects. Returns a finding for each
// project whose build task writes outside dist/; none when all is well.
func CheckBuildDirs(projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if f, ok := taskoutput.Check(p, "build", pythonBuild); ok {
			findings = append(findings, f)
		}
	}
	return findings
}
