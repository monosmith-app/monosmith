package kindchange

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/xmltext"
)

const propertyGroup = "PropertyGroup"

// pathAttribute matches an attribute that names a file or a project, written
// with either kind of slash: <ProjectReference Include="..\core\Core.csproj" />.
var pathAttribute = regexp.MustCompile(`\b((?:Include|Update|Link|Project)=")([^"]*)(")`)

// solutionEntry matches a project listed in a solution file (.sln, .slnx).
var solutionEntry = regexp.MustCompile(`"([^"\r\n]*\.(?:csproj|fsproj|vbproj))"`)

func isProjectFile(name string) bool {
	ext := path.Ext(name)
	return ext == ".csproj" || ext == ".fsproj" || ext == ".vbproj"
}

func isSolutionFile(name string) bool {
	ext := path.Ext(name)
	return ext == ".sln" || ext == ".slnx"
}

// Change answers what making the project params.Project a params.To changes:
// the SmithKind in its project file, whether it is packable, and every
// project reference and solution entry that names its directory.
func Change(workspaceRoot string, params protocol.ChangeKindParams) (protocol.ChangeKindResult, error) {
	p := params.Project
	edits := fileedits.New(workspaceRoot)
	files := edits.List(p.Root, isProjectFile)
	if len(files) == 0 {
		return protocol.ChangeKindResult{}, fmt.Errorf("no project file in %s", p.Root)
	}
	manifest := files[0]
	src, _ := edits.Read(manifest)
	var result protocol.ChangeKindResult

	src = setProperty(src, "SmithKind", string(params.To), "")
	if params.To == protocol.ProjectPackage {
		if v, found := xmltext.Element(src, "IsPackable"); found && strings.EqualFold(v, "false") {
			src, _ = xmltext.RemoveElement(src, "IsPackable")
		}
		for _, missing := range missingPackageFields(src) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("nuget.org asks for %s in %s before the first release; add it", missing, manifest))
		}
	} else {
		src = setProperty(src, "IsPackable", "false", "Version")
	}
	edits.Write(manifest, src)
	result.Warnings = append(result.Warnings, programWarnings(edits, src, manifest, p, params.To)...)

	if params.NewRoot != p.Root {
		moveOwnFiles(edits, params)
		retargetDependents(edits, params)
		retargetSolutions(edits, params)
	}
	result.Files = edits.Edits()
	return result, nil
}

func setProperty(src, name, value, before string) string {
	out, ok := xmltext.SetElementBefore(src, propertyGroup, name, value, before)
	if !ok {
		return src
	}
	return out
}

// missingPackageFields are what nuget.org asks of a package that the project
// file does not set.
func missingPackageFields(src string) []string {
	var missing []string
	if _, ok := xmltext.Element(src, "Description"); !ok {
		missing = append(missing, "a description")
	}
	_, expression := xmltext.Element(src, "PackageLicenseExpression")
	_, file := xmltext.Element(src, "PackageLicenseFile")
	if !expression && !file {
		missing = append(missing, "a license")
	}
	return missing
}

// programWarnings say what the kind change makes of the project's program.
func programWarnings(edits *fileedits.Set, src, manifest string, p protocol.Project, to protocol.ProjectKind) []string {
	output, _ := xmltext.Element(src, "OutputType")
	_, hasProgramFile := edits.Read(path.Join(p.Root, "Program.cs"))
	if !hasProgramFile {
		_, hasProgramFile = edits.Read(path.Join(p.Root, "Program.fs"))
	}
	runs := output == "Exe" || output == "WinExe"
	switch {
	case to == protocol.ProjectApp && !runs && hasProgramFile:
		return []string{fmt.Sprintf("%s has a Program file but is not built as a program: add <OutputType>Exe</OutputType> before running it", manifest)}
	case to == protocol.ProjectApp && !runs:
		return []string{"an app runs a program, and this project has no Program.cs: add one, and <OutputType>Exe</OutputType>, before running it"}
	case to != protocol.ProjectApp && runs:
		return []string{fmt.Sprintf("this project builds a program (OutputType %s); as a %s it is referenced by other projects and the program is kept but not deployed", output, to)}
	}
	return nil
}

// mapAttributes rewrites the path of every path attribute in text with fn.
// fn works on forward slashes; a path written with backslashes stays so.
func mapAttributes(text string, fn func(ref string) string) string {
	return pathAttribute.ReplaceAllStringFunc(text, func(m string) string {
		parts := pathAttribute.FindStringSubmatch(m)
		return parts[1] + restoreSlashes(parts[2], fn(strings.ReplaceAll(parts[2], `\`, "/"))) + parts[3]
	})
}

func restoreSlashes(original, changed string) string {
	if strings.Contains(original, `\`) {
		return strings.ReplaceAll(changed, "/", `\`)
	}
	return changed
}

// moveOwnFiles keeps the project's own references to other projects and shared
// files (its project file, the tests project beside it) pointing at the right
// places once its directory moves.
func moveOwnFiles(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range edits.Walk(p.Root, isProjectFile) {
		text, _ := edits.Read(file)
		dir := path.Dir(file)
		edits.Write(file, mapAttributes(text, func(ref string) string {
			return pathrefs.RebaseFrom(ref, dir, p.Root, params.NewRoot)
		}))
	}
}

// retargetDependents points every other project's references at the new
// directory.
func retargetDependents(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, other := range params.Projects {
		if other.Language != "dotnet" || other.Root == p.Root {
			continue
		}
		for _, file := range edits.Walk(other.Root, isProjectFile) {
			text, _ := edits.Read(file)
			dir := path.Dir(file)
			edits.Write(file, mapAttributes(text, func(ref string) string {
				return pathrefs.Follow(ref, dir, p.Root, params.NewRoot)
			}))
		}
	}
}

// retargetSolutions updates the solution files at the workspace root.
func retargetSolutions(edits *fileedits.Set, params protocol.ChangeKindParams) {
	p := params.Project
	for _, file := range edits.List("", isSolutionFile) {
		text, _ := edits.Read(file)
		edits.Write(file, solutionEntry.ReplaceAllStringFunc(text, func(m string) string {
			entry := solutionEntry.FindStringSubmatch(m)[1]
			norm := strings.TrimPrefix(strings.ReplaceAll(entry, `\`, "/"), "./")
			if !strings.HasPrefix(norm, p.Root+"/") {
				return m
			}
			return `"` + restoreSlashes(entry, params.NewRoot+strings.TrimPrefix(norm, p.Root)) + `"`
		}))
	}
}
