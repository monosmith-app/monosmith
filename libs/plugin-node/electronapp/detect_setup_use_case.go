package electronapp

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// forgeConfigs are the configuration files Electron Forge reads when
// package.json has no config.forge, in the order smith looks for them.
var forgeConfigs = []string{"forge.config.js", "forge.config.cjs", "forge.config.mjs", "forge.config.ts"}

// builderConfigs are the configuration files of electron-builder smith looks
// for when package.json has no "build" key. Only the JSON and YAML ones are
// read; the others make the folder unknown.
var builderConfigs = []string{
	"electron-builder.json", "electron-builder.yml", "electron-builder.yaml",
	"electron-builder.json5", "electron-builder.toml",
	"electron-builder.config.js", "electron-builder.config.cjs", "electron-builder.config.mjs", "electron-builder.config.ts",
}

var (
	// outDirString is a Forge outDir set to a string literal.
	outDirString = regexp.MustCompile("outDir\\s*:\\s*['\"`]([^'\"`$]+)['\"`]")
	// exportedObject opens the object a Forge configuration file exports.
	exportedObject = regexp.MustCompile(`(module\.exports\s*=\s*\{|export\s+default\s+\{|const\s+config(\s*:\s*ForgeConfig)?\s*=\s*\{)`)
	// yamlDirectories is a top-level directories key; yamlOutput its output.
	yamlDirectories = regexp.MustCompile(`(?m)^directories\s*:`)
	yamlOutput      = regexp.MustCompile(`(?m)^directories\s*:\s*\n(?:[ \t]+.*\n)*?[ \t]+output\s*:\s*['"]?([^'"\n#]+?)['"]?\s*(?:#.*)?$`)
)

// Detect reads the Electron setup of the project in projectDir.
//
// The app smith makes has scripts/assemble.mjs beside @electron-forge/cli.
// Another project with @electron-forge/cli is Forge; one with electron-builder
// is Builder. A project without either is None, as is one whose package.json
// cannot be read.
//
// Parameters: projectDir is the absolute project folder.
// Returns the setup.
func Detect(projectDir string) Setup {
	raw, err := os.ReadFile(filepath.Join(projectDir, "package.json"))
	if err != nil {
		return Setup{}
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return Setup{}
	}
	switch {
	case depends(doc, "@electron-forge/cli"):
		if exists(filepath.Join(projectDir, "scripts", "assemble.mjs")) {
			return Setup{Tool: SmithForge}
		}
		return forgeSetup(projectDir, doc)
	case depends(doc, "electron-builder"):
		return builderSetup(projectDir, doc)
	}
	return Setup{}
}

// forgeSetup reads Forge's outDir: package.json config.forge as an object,
// else the file config.forge names, else forge.config.*. The default is out.
func forgeSetup(projectDir string, doc *jsondoc.Document) Setup {
	s := Setup{Tool: Forge, OutDir: "out", Config: forgeConfigs[0]}
	if config := doc.Object("config", false); config != nil {
		if forge := config.Object("forge", false); forge != nil {
			s.Config, s.Editable = "package.json", true
			if v, ok := forge.String("outDir"); ok {
				s.OutDir = cleanOut(v)
			}
			return s
		}
		if file, ok := config.String("forge"); ok {
			return forgeFile(projectDir, path.Clean(filepath.ToSlash(file)), s)
		}
	}
	for _, name := range forgeConfigs {
		if exists(filepath.Join(projectDir, name)) {
			return forgeFile(projectDir, name, s)
		}
	}
	// No configuration at all: Forge's defaults. smith does not write a file
	// the person does not have.
	s.Config = ""
	return s
}

// forgeFile reads the outDir of a Forge configuration file. smith edits it
// only when the file has one outDir string, or none and one plain exported
// object to add it to.
func forgeFile(projectDir, name string, s Setup) Setup {
	s.Config = name
	raw, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(name)))
	if err != nil {
		s.OutDir = ""
		return s
	}
	text := string(raw)
	matches := outDirString.FindAllStringSubmatch(text, -1)
	switch {
	case len(matches) == 1:
		s.OutDir, s.Editable = cleanOut(matches[0][1]), true
	case len(matches) > 1 || strings.Contains(text, "outDir"):
		s.OutDir = "" // set some other way: smith cannot tell
	default:
		s.Editable = len(exportedObject.FindAllString(text, -1)) == 1
	}
	return s
}

// builderSetup reads electron-builder's directories.output: package.json
// "build", else electron-builder.json or .yml. The default is dist.
func builderSetup(projectDir string, doc *jsondoc.Document) Setup {
	s := Setup{Tool: Builder, OutDir: "dist", Config: "package.json"}
	if build := doc.Object("build", false); build != nil {
		s.Editable = true
		if dirs := build.Object("directories", false); dirs != nil {
			if v, ok := dirs.String("output"); ok {
				s.OutDir = cleanOut(v)
			}
		}
		return s
	}
	for _, name := range builderConfigs {
		file := filepath.Join(projectDir, name)
		if !exists(file) {
			continue
		}
		s.Config = name
		raw, err := os.ReadFile(file)
		if err != nil {
			s.OutDir = ""
			return s
		}
		switch path.Ext(name) {
		case ".json":
			d, err := jsondoc.ParseDocument(raw)
			if err != nil {
				s.OutDir = ""
				return s
			}
			s.Editable = true
			if dirs := d.Object("directories", false); dirs != nil {
				if v, ok := dirs.String("output"); ok {
					s.OutDir = cleanOut(v)
				}
			}
		case ".yml", ".yaml":
			if m := yamlOutput.FindStringSubmatch(string(raw)); m != nil {
				s.OutDir = cleanOut(m[1])
			}
			s.Editable = !yamlDirectories.Match(raw)
		default:
			s.OutDir = ""
		}
		return s
	}
	// No configuration file: package.json takes a new "build" key.
	s.Editable = true
	return s
}

// cleanOut is a configured folder in slash form, or "" when it holds a macro
// or is absolute, which smith cannot follow.
func cleanOut(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || strings.Contains(v, "$") || filepath.IsAbs(v) || strings.HasPrefix(v, "/") {
		return ""
	}
	return path.Clean(filepath.ToSlash(v))
}

// depends reports whether package.json names the package in dependencies or
// devDependencies.
func depends(doc *jsondoc.Document, name string) bool {
	for _, key := range []string{"devDependencies", "dependencies"} {
		if deps := doc.Object(key, false); deps != nil && deps.Has(name) {
			return true
		}
	}
	return false
}

// exists reports whether the file is there.
func exists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}
