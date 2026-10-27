package discovery

import (
	"encoding/xml"
	"path"
	"strings"
)

// MSBuildProject is what smith reads from a project file.
type MSBuildProject struct {
	// Name is the file's base name without extension (Acme.Core).
	Name string
	// PackageID is what the project is known by: PackageId, AssemblyName or
	// the file name.
	PackageID string
	// AssemblyName is the name of the program an app builds: AssemblyName, or
	// the file's base name.
	AssemblyName string
	Version      string
	// OutputType Exe makes an app.
	OutputType string
	// Kind and Stack are the SmithKind and SmithStack properties.
	Kind, Stack string
	// Test is a tests project (IsTestProject, or one that references
	// Microsoft.NET.Test.Sdk).
	Test bool
	// References are the project files it references, as written.
	References []string
}

type msbuild struct {
	PropertyGroups []struct {
		PackageID     string `xml:"PackageId"`
		AssemblyName  string `xml:"AssemblyName"`
		Version       string `xml:"Version"`
		OutputType    string `xml:"OutputType"`
		IsTestProject string `xml:"IsTestProject"`
		SmithKind     string `xml:"SmithKind"`
		SmithStack    string `xml:"SmithStack"`
	} `xml:"PropertyGroup"`
	ItemGroups []struct {
		ProjectReferences []struct {
			Include string `xml:"Include,attr"`
		} `xml:"ProjectReference"`
		PackageReferences []struct {
			Include string `xml:"Include,attr"`
		} `xml:"PackageReference"`
	} `xml:"ItemGroup"`
}

// ParseProject reads a project file; file is its workspace-relative path.
func ParseProject(file string, src []byte) (MSBuildProject, error) {
	var m msbuild
	if err := xml.Unmarshal(src, &m); err != nil {
		return MSBuildProject{}, err
	}
	base := path.Base(file)
	p := MSBuildProject{Name: strings.TrimSuffix(base, path.Ext(base))}
	assembly := ""
	for _, g := range m.PropertyGroups {
		first := func(dst *string, v string) {
			if *dst == "" && strings.TrimSpace(v) != "" {
				*dst = strings.TrimSpace(v)
			}
		}
		first(&p.PackageID, g.PackageID)
		first(&assembly, g.AssemblyName)
		first(&p.Version, g.Version)
		first(&p.OutputType, g.OutputType)
		first(&p.Kind, g.SmithKind)
		first(&p.Stack, g.SmithStack)
		if strings.EqualFold(strings.TrimSpace(g.IsTestProject), "true") {
			p.Test = true
		}
	}
	p.AssemblyName = assembly
	if p.AssemblyName == "" {
		p.AssemblyName = p.Name
	}
	if p.PackageID == "" {
		p.PackageID = assembly
	}
	if p.PackageID == "" {
		p.PackageID = p.Name
	}
	for _, g := range m.ItemGroups {
		for _, r := range g.ProjectReferences {
			p.References = append(p.References, strings.ReplaceAll(r.Include, `\`, "/"))
		}
		for _, r := range g.PackageReferences {
			if r.Include == "Microsoft.NET.Test.Sdk" {
				p.Test = true
			}
		}
	}
	return p, nil
}
