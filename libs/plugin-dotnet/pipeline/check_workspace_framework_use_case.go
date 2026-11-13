package pipeline

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckWorkspaceFramework warns when projects repeat their target framework
// instead of inheriting it from Directory.Build.props, the way new projects
// do. The fix moves the framework to the props file and removes it from the
// projects that share it. It is never safe: it edits project files the user
// owns, so doctor asks first. It reports nothing when no project sets a
// single target framework of its own.
func CheckWorkspaceFramework(workspaceRoot string, projects []protocol.Project) (protocol.Finding, bool) {
	files := map[string]string{}
	for _, p := range projects {
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		for _, pattern := range []string{"*.csproj", filepath.Join("tests", "*.csproj")} {
			matches, _ := filepath.Glob(filepath.Join(dir, pattern))
			for _, m := range matches {
				raw, err := os.ReadFile(m)
				if err != nil {
					continue
				}
				rel, err := filepath.Rel(workspaceRoot, m)
				if err != nil {
					continue
				}
				files[filepath.ToSlash(rel)] = string(raw)
			}
		}
	}
	propsRaw, _ := os.ReadFile(filepath.Join(workspaceRoot, propsFile))
	props, changed, framework := scaffolding.ShareFramework(string(propsRaw), files)
	if framework == "" {
		return protocol.Finding{}, false
	}
	paths := make([]string, 0, len(changed))
	for path := range changed {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	edits := []protocol.FileEdit{{Path: propsFile, Content: props}}
	for _, path := range paths {
		edits = append(edits, protocol.FileEdit{Path: path, Content: changed[path]})
	}
	return protocol.Finding{
		Check: "projects inherit their target framework from the workspace", Area: "config", Warning: true,
		Problem: "projects set their own target framework, so one edit of the workspace does not reach them",
		Detail:  "New projects take it from Directory.Build.props. These project files set " + framework + " themselves: " + joinPaths(paths),
		Remedy:  "move <TargetFramework>" + framework + "</TargetFramework> to Directory.Build.props and delete it from the project files",
		Fix:     &protocol.Fix{Description: "share " + framework + " through Directory.Build.props", Safe: false, Files: edits},
	}, true
}

func joinPaths(paths []string) string {
	out := ""
	for i, p := range paths {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
