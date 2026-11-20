package workspacejoin

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/dependencies"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// solutionEntry matches a project a solution file lists (.sln, .slnx).
var solutionEntry = regexp.MustCompile(`"([^"\r\n]*\.(?:csproj|fsproj|vbproj))"`)

// Solutions are the solution files at the workspace root, sorted.
//
// Parameters: workspaceRoot is the workspace folder.
// Returns: the base names of the .sln and .slnx files there.
func Solutions(workspaceRoot string) []string {
	entries, _ := os.ReadDir(workspaceRoot)
	var out []string
	for _, e := range entries {
		ext := path.Ext(e.Name())
		if !e.IsDir() && (ext == ".sln" || ext == ".slnx") {
			out = append(out, e.Name())
		}
	}
	slices.Sort(out)
	return out
}

// Listed are the project files a solution lists, workspace-relative with
// forward slashes.
//
// Parameters: text is the solution file's content.
// Returns: the cleaned paths.
func Listed(text string) []string {
	var out []string
	for _, m := range solutionEntry.FindAllStringSubmatch(text, -1) {
		out = append(out, path.Clean(strings.ReplaceAll(m[1], `\`, "/")))
	}
	return out
}

// Join says, for each .NET project, whether a solution at the workspace root
// lists its project file. A workspace without a solution has no .NET
// workspace: smith restores every project on its own, so each is reached and
// Join reports none. A project no solution lists gets the dotnet sln add
// command for the solution, or a choice of solution when there are several.
// It writes nothing.
//
// Parameters: workspaceRoot is the workspace folder; params are the projects.
// Returns: one state for each project when there is a solution, else none.
func Join(workspaceRoot string, params protocol.JoinParams) (protocol.JoinResult, error) {
	res := protocol.JoinResult{Projects: []protocol.JoinState{}}
	solutions := Solutions(workspaceRoot)
	if len(solutions) == 0 {
		return res, nil
	}
	var listed []string
	for _, s := range solutions {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, s))
		if err != nil {
			return res, err
		}
		listed = append(listed, Listed(string(raw))...)
	}
	words := "the solution " + strings.Join(solutions, " or ")
	for _, p := range params.Projects {
		file, err := dependencies.ProjectFileOf(workspaceRoot, p)
		if err != nil {
			continue
		}
		joined := slices.Contains(listed, file)
		state := protocol.JoinState{Project: p.Name, Joined: joined, Workspace: words}
		if !joined {
			state.Problem = "is not in " + words + ": a build of the solution and the IDE leave it out"
			state.Fix = addFix(solutions, file, p.Name)
		}
		res.Projects = append(res.Projects, state)
	}
	return res, nil
}

// addFix is dotnet sln add for one solution, or a choice of the solution.
func addFix(solutions []string, file, project string) *protocol.Fix {
	step := func(sln string) protocol.FixStep {
		return protocol.FixStep{Description: "add " + file + " to " + sln, Argv: []string{"dotnet", "sln", sln, "add", file}}
	}
	fix := &protocol.Fix{Description: "add the project to the solution with dotnet sln add", Safe: true, AppliesTo: []string{project}}
	if len(solutions) == 1 {
		fix.Steps = []protocol.FixStep{step(solutions[0])}
		return fix
	}
	choice := protocol.FixChoice{Question: "Which solution does " + project + " join?"}
	for _, sln := range solutions {
		choice.Options = append(choice.Options, protocol.FixChoiceOption{Label: sln, Fix: protocol.Fix{Description: "add it to " + sln, Steps: []protocol.FixStep{step(sln)}}})
	}
	fix.Choices = []protocol.FixChoice{choice}
	return fix
}
