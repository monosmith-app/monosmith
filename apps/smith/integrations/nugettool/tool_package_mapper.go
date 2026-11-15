package nugettool

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// rids are the .NET runtime identifiers a static Go binary for each os/arch
// runs under.
var rids = map[string]string{
	"linux/amd64":   "linux-x64",
	"linux/arm64":   "linux-arm64",
	"darwin/amd64":  "osx-x64",
	"darwin/arm64":  "osx-arm64",
	"windows/amd64": "win-x64",
	"windows/arm64": "win-arm64",
}

// RID is the runtime identifier of an os/arch.
func RID(goos, goarch string) (string, bool) {
	rid, ok := rids[goos+"/"+goarch]
	return rid, ok
}

// FileName is a package's file name as NuGet spells it.
func FileName(id, version string) string {
	return strings.ToLower(id) + "." + version + ".nupkg"
}

// Metadata describes the tool on NuGet.
type Metadata struct {
	ID, Version, Authors, Description, License, ProjectURL, Repository string
	// Readme is the file name of the package's readme, or "" for none.
	Readme string
}

// packageType is DotnetTool for the package users install and
// DotnetToolRidPackage for the per-platform packages it points at.
func nuspec(m Metadata, packageType string) string {
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<package xmlns=\"http://schemas.microsoft.com/packaging/2012/06/nuspec.xsd\">\n  <metadata>\n")
	element(&b, "id", m.ID)
	element(&b, "version", m.Version)
	element(&b, "authors", m.Authors)
	if m.License != "" {
		fmt.Fprintf(&b, "    <license type=\"expression\">%s</license>\n", escape(m.License))
		// nuget.org refuses a license expression without the URL older
		// clients show, as dotnet pack writes it.
		element(&b, "licenseUrl", "https://licenses.nuget.org/"+m.License)
	}
	if m.ProjectURL != "" {
		element(&b, "projectUrl", m.ProjectURL)
	}
	if m.Readme != "" {
		element(&b, "readme", m.Readme)
	}
	element(&b, "description", m.Description)
	fmt.Fprintf(&b, "    <packageTypes>\n      <packageType name=\"%s\" />\n    </packageTypes>\n", packageType)
	if m.Repository != "" {
		fmt.Fprintf(&b, "    <repository type=\"git\" url=\"%s\" />\n", escape(m.Repository))
	}
	b.WriteString("  </metadata>\n</package>")
	return b.String()
}

// toolSettings is the tool package's DotnetToolSettings.xml: the command and
// the package holding it for each runtime identifier.
func toolSettings(command string, ridPackages [][2]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<DotNetCliTool Version=\"2\">\n  <Commands>\n    <Command Name=\"%s\" />\n  </Commands>\n  <RuntimeIdentifierPackages>\n", escape(command))
	for _, p := range ridPackages {
		fmt.Fprintf(&b, "    <RuntimeIdentifierPackage RuntimeIdentifier=\"%s\" Id=\"%s\" />\n", p[0], escape(p[1]))
	}
	b.WriteString("  </RuntimeIdentifierPackages>\n</DotNetCliTool>")
	return b.String()
}

// ridToolSettings is a platform package's DotnetToolSettings.xml: the
// command runs the binary itself, with no .NET runtime in between.
func ridToolSettings(command, entryPoint string) string {
	return fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<DotNetCliTool Version=\"2\">\n  <Commands>\n    <Command Name=\"%s\" EntryPoint=\"%s\" Runner=\"executable\" />\n  </Commands>\n</DotNetCliTool>", escape(command), escape(entryPoint))
}

// contentTypes is the package's [Content_Types].xml (Open Packaging
// Conventions): a type per extension, and an override for each file
// without one, such as a Unix binary.
func contentTypes(files []string) string {
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<Types xmlns=\"http://schemas.openxmlformats.org/package/2006/content-types\">\n")
	b.WriteString("  <Default Extension=\"rels\" ContentType=\"application/vnd.openxmlformats-package.relationships+xml\" />\n")
	b.WriteString("  <Default Extension=\"psmdcp\" ContentType=\"application/vnd.openxmlformats-package.core-properties+xml\" />\n")
	seen := map[string]bool{"rels": true, "psmdcp": true}
	var overrides []string
	for _, f := range files {
		base := f[strings.LastIndex(f, "/")+1:]
		dot := strings.LastIndex(base, ".")
		if dot <= 0 {
			overrides = append(overrides, f)
			continue
		}
		ext := base[dot+1:]
		if !seen[ext] {
			seen[ext] = true
			fmt.Fprintf(&b, "  <Default Extension=\"%s\" ContentType=\"application/octet\" />\n", escape(ext))
		}
	}
	for _, f := range overrides {
		fmt.Fprintf(&b, "  <Override PartName=\"/%s\" ContentType=\"application/octet\" />\n", escape(f))
	}
	b.WriteString("</Types>")
	return b.String()
}

// relationships is the package's _rels/.rels: where its manifest and core
// properties are.
func relationships(nuspecFile, propertiesFile string) string {
	return fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\">\n  <Relationship Type=\"http://schemas.microsoft.com/packaging/2010/07/manifest\" Target=\"/%s\" Id=\"R1\" />\n  <Relationship Type=\"http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties\" Target=\"/%s\" Id=\"R2\" />\n</Relationships>", escape(nuspecFile), propertiesFile)
}

// coreProperties repeats the manifest's identity as OPC core properties.
func coreProperties(m Metadata) string {
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<coreProperties xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:dcterms=\"http://purl.org/dc/terms/\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns=\"http://schemas.openxmlformats.org/package/2006/metadata/core-properties\">\n")
	fmt.Fprintf(&b, "  <dc:creator>%s</dc:creator>\n  <dc:description>%s</dc:description>\n  <dc:identifier>%s</dc:identifier>\n  <version>%s</version>\n  <keywords></keywords>\n  <lastModifiedBy>smith</lastModifiedBy>\n</coreProperties>",
		escape(m.Authors), escape(m.Description), escape(m.ID), escape(m.Version))
	return b.String()
}

func element(b *strings.Builder, name, value string) {
	fmt.Fprintf(b, "    <%s>%s</%s>\n", name, escape(value), name)
}

func escape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
