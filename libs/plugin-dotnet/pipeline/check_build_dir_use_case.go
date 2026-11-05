package pipeline

import (
	"encoding/xml"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/buildrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/xmltext"
)

// artifactsDir is the folder smith builds every .NET project into: the
// artifacts layout of MSBuild, one folder for the workspace.
var artifactsDir = protocol.BuildDir + "/dotnet"

const propsFile = "Directory.Build.props"

var (
	// outputProperties are the properties that name where a project builds
	// itself. An explicit one wins over ArtifactsPath.
	outputProperties = regexp.MustCompile(`<(OutputPath|BaseOutputPath|IntermediateOutputPath|BaseIntermediateOutputPath)\b`)
	hintPath         = regexp.MustCompile(`<HintPath>\s*([^<]+?)\s*</HintPath>`)
	importElement    = regexp.MustCompile(`<Import\b`)
)

// dotnetGroup is the projects that one Directory.Build.props governs.
type dotnetGroup struct {
	// props is the workspace-relative file; it may not exist yet.
	props    string
	exists   bool
	projects []protocol.Project
	// artifacts is the ArtifactsPath it sets, resolved from the workspace
	// root ("" when it sets none).
	artifacts string
}

// CheckBuildDir reports where the .NET projects build, from what MSBuild
// reads: the ArtifactsPath of the nearest Directory.Build.props (MSBuild reads
// the first one it finds going up, and no other), else bin/ and obj/ in each
// project. A project file that sets OutputPath or BaseOutputPath itself is
// reported too, but it is never rewritten: those win over the artifacts path.
//
// A Directory.Build.props that imports another file, or sets ArtifactsPath
// from a property smith cannot evaluate, is left unchecked.
//
// Parameters: workspaceRoot is the absolute workspace folder; projects the
// .NET projects. Returns one finding for each props file, and false when
// there is none to report.
func CheckBuildDir(workspaceRoot string, projects []protocol.Project) ([]protocol.Finding, bool) {
	groups := map[string]*dotnetGroup{}
	var order []string
	for _, p := range projects {
		if len(projectFiles(workspaceRoot, p)) == 0 {
			continue // no project file to build: nothing to report
		}
		props, exists := nearestProps(workspaceRoot, p.Root)
		g, ok := groups[props]
		if !ok {
			g = &dotnetGroup{props: props, exists: exists}
			groups[props] = g
			order = append(order, props)
		}
		g.projects = append(g.projects, p)
	}
	sort.Strings(order)
	var findings []protocol.Finding
	for _, key := range order {
		if f, ok := groupFinding(workspaceRoot, groups[key]); ok {
			findings = append(findings, f)
		}
	}
	return findings, len(findings) > 0
}

// nearestProps is the Directory.Build.props that governs a project: the first
// one going up from its folder. When there is none it is the file smith would
// write at the workspace root.
func nearestProps(workspaceRoot, root string) (string, bool) {
	for dir := root; ; dir = path.Dir(dir) {
		rel := path.Join(dir, propsFile)
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err == nil {
			return rel, true
		}
		if dir == "." {
			return propsFile, false
		}
	}
}

// projectFiles lists the project files of a project and of its tests folder.
func projectFiles(workspaceRoot string, p protocol.Project) []string {
	var out []string
	for _, dir := range []string{p.Root, path.Join(p.Root, "tests")} {
		entries, _ := os.ReadDir(filepath.Join(workspaceRoot, filepath.FromSlash(dir)))
		for _, e := range entries {
			switch strings.ToLower(path.Ext(e.Name())) {
			case ".csproj", ".fsproj", ".vbproj":
				out = append(out, path.Join(dir, e.Name()))
			}
		}
	}
	return out
}

func read(workspaceRoot, rel string) string {
	raw, _ := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return string(raw)
}

func groupFinding(workspaceRoot string, g *dotnetGroup) (protocol.Finding, bool) {
	var names []string
	for _, p := range g.projects {
		names = append(names, p.Name)
	}
	check := ".NET builds into " + artifactsDir
	if g.props != propsFile {
		check = strings.Join(names, ", ") + " builds into " + artifactsDir
	}
	propsDir := path.Dir(g.props)
	var src string
	if g.exists {
		src = read(workspaceRoot, g.props)
		if importElement.MatchString(src) {
			return protocol.Finding{}, false // chained to another props file: not read
		}
		if v, ok := xmltext.Element(src, "ArtifactsPath"); ok {
			if !strings.HasPrefix(v, "$(MSBuildThisFileDirectory)") {
				return protocol.Finding{}, false // a property smith cannot evaluate
			}
			g.artifacts = path.Clean(path.Join(propsDir, strings.TrimPrefix(v, "$(MSBuildThisFileDirectory)")))
		}
	}
	explicit := explicitOutputs(workspaceRoot, g, src)
	if g.artifacts != "" && protocol.InBuildDir(g.artifacts) && len(explicit) == 0 {
		return protocol.Finding{Check: check, OK: true}, true
	}
	where := "bin/ and obj/ of each project"
	if g.artifacts != "" {
		where = g.artifacts + "/"
	}
	f := protocol.Finding{
		Check:  check,
		Detail: strings.Join(names, ", ") + verb(len(names)) + " into " + where + ", not into " + artifactsDir + "/",
		Remedy: "set <ArtifactsPath>" + propsValue(propsDir) + "</ArtifactsPath> in " + g.props + ", or run smith doctor --fix",
	}
	if refusal := dotnetRefusal(workspaceRoot, g, explicit); refusal != "" {
		f.Warning = true
		f.Remedy = "smith does not edit " + g.props + " for you, because " + refusal + ". To move it by hand: " + strings.TrimSuffix(f.Remedy, ", or run smith doctor --fix")
		return f, true
	}
	edited, refusal := withArtifactsPath(src, g.exists, propsValue(propsDir))
	if refusal != "" {
		f.Warning = true
		f.Remedy = "smith does not edit " + g.props + " for you, because " + refusal + ". To move it by hand: " + strings.TrimSuffix(f.Remedy, ", or run smith doctor --fix")
		return f, true
	}
	move := protocol.Fix{Description: "Build the .NET projects into " + artifactsDir, Files: []protocol.FileEdit{{Path: g.props, Content: edited}}}
	f.Fix = &protocol.Fix{
		Description: "Build " + strings.Join(names, ", ") + " into " + artifactsDir,
		AppliesTo:   names,
		Choices: []protocol.FixChoice{{
			Question: strings.Join(names, ", ") + verb(len(names)) + " into " + where + ". smith builds every .NET project into " + artifactsDir + " (the artifacts layout of MSBuild). Moving it sets ArtifactsPath in " + g.props + " and keeps every other line. The old bin/ and obj/ folders stay on disk, unused (git ignores them), and nothing is committed. The next screen shows every changed line.",
			Options: []protocol.FixChoiceOption{
				{Label: "build into " + artifactsDir + " (edits " + g.props + ")", Fix: move},
				{Label: "leave it as it is (nothing changes)"},
			},
		}},
	}
	return f, true
}

// propsValue is the ArtifactsPath value as the props file in dir writes it.
func propsValue(dir string) string {
	up := ""
	if dir != "." {
		up = strings.Repeat("../", strings.Count(dir, "/")+1)
	}
	return "$(MSBuildThisFileDirectory)" + up + artifactsDir
}

// explicitOutputs names the files of the group that set where a project
// builds itself.
func explicitOutputs(workspaceRoot string, g *dotnetGroup, props string) []string {
	var out []string
	if outputProperties.MatchString(props) {
		out = append(out, g.props)
	}
	for _, p := range g.projects {
		for _, f := range projectFiles(workspaceRoot, p) {
			if outputProperties.MatchString(read(workspaceRoot, f)) {
				out = append(out, f)
			}
		}
	}
	return out
}

// dotnetRefusal says why the build cannot be moved safely, or "".
func dotnetRefusal(workspaceRoot string, g *dotnetGroup, explicit []string) string {
	if len(explicit) > 0 {
		return strings.Join(explicit, ", ") + " set" + plural(len(explicit)) + " OutputPath or BaseOutputPath itself, which wins over ArtifactsPath"
	}
	old := map[string]string{} // a project's old folder, from the workspace root, to its name
	for _, p := range g.projects {
		old[path.Join(p.Root, "bin")] = p.Name
		old[path.Join(p.Root, "obj")] = p.Name
		if mentioned := buildrefs.Mentions(workspaceRoot, p.Root, "bin", "obj"); len(mentioned) > 0 {
			return strings.Join(mentioned, ", ") + " name" + plural(len(mentioned)) + " the bin/ or obj/ folder of " + p.Name + " and would stop finding the build"
		}
	}
	for _, p := range g.projects {
		for _, f := range projectFiles(workspaceRoot, p) {
			for _, m := range hintPath.FindAllStringSubmatch(read(workspaceRoot, f), -1) {
				target := path.Clean(path.Join(path.Dir(f), strings.ReplaceAll(m[1], `\`, "/")))
				for dir, name := range old {
					if target == dir || strings.HasPrefix(target, dir+"/") {
						return f + " references a built file of " + name + " by its path (" + m[1] + "); that path would stop existing"
					}
				}
			}
		}
	}
	return ""
}

func plural(n int) string {
	if n == 1 {
		return "s"
	}
	return ""
}

// withArtifactsPath sets ArtifactsPath (and UseArtifactsOutput) in the text of
// a Directory.Build.props, keeping every other line. Without a file it is the
// smallest props file that does it. It returns a refusal when the text is not
// valid XML or the edit is not valid either.
func withArtifactsPath(src string, exists bool, value string) (string, string) {
	if !exists || strings.TrimSpace(src) == "" {
		return "<Project>\n  <!-- Every build goes to the workspace's dist/ (smith): bin/ and obj/ of every project live under dist/dotnet. -->\n  <PropertyGroup>\n    <UseArtifactsOutput>true</UseArtifactsOutput>\n    <ArtifactsPath>" + value + "</ArtifactsPath>\n  </PropertyGroup>\n</Project>\n", ""
	}
	if !validXML(src) {
		return "", "it is not valid XML, so smith cannot edit it safely"
	}
	edited := src
	if _, ok := xmltext.Element(edited, "ArtifactsPath"); ok {
		edited, _ = xmltext.SetElement(edited, "PropertyGroup", "ArtifactsPath", value)
		if _, ok := xmltext.Element(edited, "UseArtifactsOutput"); ok {
			edited, _ = xmltext.SetElement(edited, "PropertyGroup", "UseArtifactsOutput", "true")
		}
	} else {
		eol := "\n"
		if strings.Contains(src, "\r\n") {
			eol = "\r\n"
		}
		block := "  <!-- Every build goes to the workspace's dist/ (smith). -->" + eol + "  <PropertyGroup>" + eol + "    <UseArtifactsOutput>true</UseArtifactsOutput>" + eol + "    <ArtifactsPath>" + value + "</ArtifactsPath>" + eol + "  </PropertyGroup>" + eol
		i := strings.LastIndex(edited, "</Project>")
		if i < 0 {
			return "", "it has no closing </Project>, so smith cannot edit it safely"
		}
		lead := ""
		if i > 0 && edited[i-1] != '\n' {
			lead = eol
		}
		edited = edited[:i] + lead + block + edited[i:]
	}
	if !validXML(edited) {
		return "", "the edit would not be valid XML"
	}
	return edited, ""
}

func validXML(src string) bool {
	d := xml.NewDecoder(strings.NewReader(src))
	for {
		if _, err := d.Token(); err == io.EOF {
			return true
		} else if err != nil {
			return false
		}
	}
}

// verb is "builds" for one project and "build" for several, after a name list.
func verb(n int) string {
	if n == 1 {
		return " builds"
	}
	return " build"
}
