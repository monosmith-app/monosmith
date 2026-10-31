package discovery

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Shape keys the node plugin reports for a project (see protocol.Project.Shape).
const (
	ShapeBuildTool     = "build.tool"
	ShapeTestRunner    = "test.runner"
	ShapeExportsTarget = "exports.target"
	ShapeTsconfigFiles = "tsconfig.files"
	ShapeLintConfig    = "lint.config"
	ShapeScripts       = "scripts"
	ShapeLayout        = "layout"
	// ShapeBuildOutput says where the build script writes: ShapeOutputProject
	// (the project's own folder), ShapeOutputWorkspace (a folder outside the
	// project, the workspace's dist/<project dir>) or none.
	ShapeBuildOutput = "build.output"
	// ShapeExportsMap is ShapeMapped when package.json has an exports map, else none.
	ShapeExportsMap = "exports.map"
	// ShapeJestConfig says where Jest is configured: ShapeConfigFile (a
	// jest.config file), ShapeConfigPackage (the "jest" field) or none.
	ShapeJestConfig = "jest.config"
)

// Shape values the plugins that follow a shape also write.
const (
	ShapeNone       = "none"
	ShapeSwcJest    = "@swc/jest"
	ShapeTsJest     = "ts-jest"
	ShapeTargetDist = "dist"
	ShapeTargetSrc  = "src"
	// ShapeOutputProject and ShapeOutputWorkspace are the values of ShapeBuildOutput.
	ShapeOutputProject   = "project"
	ShapeOutputWorkspace = "workspace"
	// ShapeMapped is the value of ShapeExportsMap for a project with an exports map.
	ShapeMapped = "map"
	// ShapeConfigFile and ShapeConfigPackage are the values of ShapeJestConfig.
	ShapeConfigFile    = "file"
	ShapeConfigPackage = "package"
	shapeOther         = "other"
	shapeProject       = "project"
	shapeWorkspace     = "root"
)

var (
	tsconfigFile = regexp.MustCompile(`^tsconfig.*\.json$`)
	eslintFile   = regexp.MustCompile(`^(eslint\.config\.[cm]?[jt]s|\.eslintrc(\..+)?)$`)
	// layoutIgnored are folders that hold output or tools, not the layout.
	layoutIgnored = map[string]bool{"node_modules": true, "dist": true, "coverage": true, "build": true}
)

// ProjectShape describes how the project in dir is made, as the facts the
// host compares between sibling projects: the build tool, the test runner,
// where the entry points lead, the tsconfig files, the lint configuration,
// the script names and the top-level folders. raw is the project's
// package.json and pkg its parsed subset. Every key is present; a fact that
// does not apply is "none".
func ProjectShape(dir string, raw []byte, pkg PackageManifest) map[string]string {
	entries, _ := os.ReadDir(dir)
	var tsconfigs, folders []string
	lint := shapeWorkspace
	for _, e := range entries {
		name := e.Name()
		switch {
		case e.IsDir():
			if !layoutIgnored[name] && !strings.HasPrefix(name, ".") {
				folders = append(folders, name)
			}
		case tsconfigFile.MatchString(name):
			tsconfigs = append(tsconfigs, name)
		case eslintFile.MatchString(name):
			lint = shapeProject
		}
	}
	scripts := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		scripts = append(scripts, name)
	}
	sort.Strings(scripts)
	return map[string]string{
		ShapeBuildTool:     scriptProgram(pkg.Scripts["build"]),
		ShapeTestRunner:    testRunner(dir, raw, pkg.Scripts["test"]),
		ShapeExportsTarget: exportsTarget(raw),
		ShapeTsconfigFiles: joinOrNone(tsconfigs),
		ShapeLintConfig:    lint,
		ShapeScripts:       joinOrNone(scripts),
		ShapeLayout:        joinOrNone(folders),
		ShapeBuildOutput:   buildOutput(pkg.Scripts["build"]),
		ShapeExportsMap:    exportsMap(raw),
		ShapeJestConfig:    jestConfigPlace(dir, raw),
	}
}

// outputFlag matches the flags of a build script that name a folder or file it writes.
var outputFlag = regexp.MustCompile(`--(?:outDir|outdir|outfile|declarationDir)[= ]+"?([^"\s&;|]+)"?`)

// buildOutput says where a build script writes: "project" when every output
// flag names a place inside the project folder, "workspace" when one leads out
// of it (../../dist/packages/a), "none" when the script names none.
func buildOutput(script string) string {
	matches := outputFlag.FindAllStringSubmatch(script, -1)
	if len(matches) == 0 {
		return ShapeNone
	}
	for _, m := range matches {
		if p := path.Clean(m[1]); p == ".." || strings.HasPrefix(p, "../") {
			return ShapeOutputWorkspace
		}
	}
	return ShapeOutputProject
}

// exportsMap is "map" when package.json has an exports object, else "none".
func exportsMap(raw []byte) string {
	var fields struct {
		Exports json.RawMessage `json:"exports"`
	}
	if json.Unmarshal(raw, &fields) == nil && len(fields.Exports) > 0 && fields.Exports[0] == '{' {
		return ShapeMapped
	}
	return ShapeNone
}

// jestConfigPlace says where Jest is configured: "file" for a jest.config
// file, "package" for the "jest" field of package.json, else "none".
func jestConfigPlace(dir string, raw []byte) string {
	if matches, _ := filepath.Glob(filepath.Join(dir, "jest.config.*")); len(matches) > 0 {
		return ShapeConfigFile
	}
	var fields struct {
		Jest json.RawMessage `json:"jest"`
	}
	if json.Unmarshal(raw, &fields) == nil && len(fields.Jest) > 0 {
		return ShapeConfigPackage
	}
	return ShapeNone
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return ShapeNone
	}
	return strings.Join(values, ",")
}

// scriptProgram is the program a script starts with, "none" for no script.
func scriptProgram(script string) string {
	fields := strings.Fields(script)
	for len(fields) > 1 && (fields[0] == "npx" || fields[0] == "cross-env") {
		fields = fields[1:]
	}
	if len(fields) == 0 {
		return ShapeNone
	}
	return fields[0]
}

// testRunner names the runner of the test script. Jest is told apart by the
// transform its configuration names: a runner of the same program can compile
// with @swc/jest or ts-jest.
func testRunner(dir string, raw []byte, script string) string {
	program := scriptProgram(script)
	switch {
	case program == ShapeNone:
		return ShapeNone
	case program == "vitest":
		return "vitest"
	case program == "node" && strings.Contains(script, "--test"):
		return "node:test"
	case program != "jest":
		return program
	}
	config := jestConfig(dir, raw)
	switch {
	case strings.Contains(config, ShapeSwcJest):
		return ShapeSwcJest
	case strings.Contains(config, ShapeTsJest):
		return ShapeTsJest
	}
	return "jest"
}

// jestConfig is the text of the project's Jest configuration: the "jest" field
// of package.json and any jest.config file.
func jestConfig(dir string, raw []byte) string {
	var fields struct {
		Jest json.RawMessage `json:"jest"`
	}
	_ = json.Unmarshal(raw, &fields)
	text := string(fields.Jest)
	matches, _ := filepath.Glob(filepath.Join(dir, "jest.config.*"))
	for _, m := range matches {
		if body, err := os.ReadFile(m); err == nil {
			text += "\n" + string(body)
		}
	}
	return text
}

// exportsTarget says where the entry points of a project lead: "dist" when
// the first one is in a build folder, "src" in the source folder, "other"
// elsewhere, "none" without an entry point. The exports field comes first,
// then main and module.
func exportsTarget(raw []byte) string {
	var fields struct {
		Exports json.RawMessage `json:"exports"`
		Main    string          `json:"main"`
		Module  string          `json:"module"`
	}
	if json.Unmarshal(raw, &fields) != nil {
		return ShapeNone
	}
	entries := append(entryPaths(fields.Exports), fields.Main, fields.Module)
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		segments := strings.Split(path.Clean(entry), "/")
		for _, s := range segments {
			if s == ShapeTargetDist {
				return ShapeTargetDist
			}
		}
		if segments[0] == ShapeTargetSrc {
			return ShapeTargetSrc
		}
		return shapeOther
	}
	return ShapeNone
}

// entryPaths lists the file paths an exports value names, the keys of an
// object in sorted order and the types conditions left out.
func entryPaths(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	var out []string
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			out = append(out, t)
		case map[string]any:
			keys := make([]string, 0, len(t))
			for k := range t {
				if k != "types" {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			for _, k := range keys {
				walk(t[k])
			}
		case []any:
			for _, item := range t {
				walk(item)
			}
		}
	}
	walk(value)
	return out
}
