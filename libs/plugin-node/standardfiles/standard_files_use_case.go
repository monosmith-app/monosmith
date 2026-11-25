package standardfiles

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Reasons the files carry, for the doctor finding and the conversion summary.
const (
	reasonSource   = "With the source condition, Ctrl+click on an import of the library opens its TypeScript source, not the built .d.ts, and tsc reads the source with no build first; Node and npm still use the build."
	reasonTsconfig = "The project's tsconfig takes the workspace's standard shape, which new projects copy."
)

// sourceEntries are the entry sources a library can have, in the order
// looked for.
var sourceEntries = []string{"src/index.ts", "src/index.mts", "src/index.tsx"}

// Standard gives smith's standard version of the configuration files of the
// Node projects: the source condition in each library's "exports" and in the
// root tsconfig.base.json, and the tsconfig files of the projects whose own
// adds nothing the standard lacks (see ReplaceableTsconfig).
//
// Parameters: root is the workspace; params name the projects.
// Returns: the standard files, project by project, the workspace's base last.
func Standard(root string, params protocol.StandardParams) (protocol.StandardResult, error) {
	res := protocol.StandardResult{Files: []protocol.StandardFileContent{}}
	var sources []protocol.StandardFileContent
	testSuffix := ""
	for _, p := range params.Projects {
		dir := filepath.Join(root, filepath.FromSlash(p.Root))
		raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
		if err != nil {
			continue
		}
		if p.Kind == protocol.ProjectLib || p.Kind == protocol.ProjectPackage {
			if f, ok := libraryExports(p, dir, raw); ok {
				sources = append(sources, f)
			}
		}
		if testSuffix == "" {
			testSuffix = scaffolding.TestSuffix(root)
		}
		res.Files = append(res.Files, tsconfigFiles(p, dir, raw, testSuffix)...)
	}
	if len(sources) == 0 {
		return res, nil
	}
	raw, err := os.ReadFile(filepath.Join(root, tsconfigchain.Base))
	if err != nil {
		return res, nil
	}
	options, ok := BaseCompilerOptions(raw)
	if !ok {
		// The condition would be read by nothing: the libraries keep their
		// exports as they are.
		return res, nil
	}
	res.Files = append(res.Files, sources...)
	res.Files = append(res.Files, protocol.StandardFileContent{
		Path: tsconfigchain.Base, Mode: protocol.StandardKeys, Format: "json",
		Keys: map[string]json.RawMessage{"compilerOptions": options}, Reason: reasonSource,
	})
	return res, nil
}

// libraryExports is the "exports" with the source condition, for a library
// whose entry points lead to its build and whose source entry exists.
func libraryExports(p protocol.Project, dir string, raw []byte) (protocol.StandardFileContent, bool) {
	for _, entry := range sourceEntries {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(entry))); err != nil {
			continue
		}
		exports, ok := SourceExports(raw, "./"+entry)
		if !ok {
			return protocol.StandardFileContent{}, false
		}
		return protocol.StandardFileContent{
			Project: p.Name, Path: path.Join(p.Root, "package.json"), Mode: protocol.StandardKeys, Format: "json",
			Keys: map[string]json.RawMessage{"exports": exports}, Reason: reasonSource,
		}, true
	}
	return protocol.StandardFileContent{}, false
}

// tsconfigFiles are the standard tsconfig.json and tsconfig.build.json of a
// project whose own files can take them without changing how it compiles.
func tsconfigFiles(p protocol.Project, dir string, pkg []byte, testSuffix string) []protocol.StandardFileContent {
	if nxLayout(dir) {
		return nil
	}
	if info, err := os.Stat(filepath.Join(dir, "src")); err != nil || !info.IsDir() {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(dir, TsconfigJSON))
	if err != nil {
		return nil
	}
	var scripts struct {
		Scripts map[string]string `json:"scripts"`
	}
	_ = json.Unmarshal(pkg, &scripts)
	if !ReplaceableTsconfig(p.Root, raw, scripts.Scripts["build"]) {
		return nil
	}
	out := []protocol.StandardFileContent{{
		Project: p.Name, Path: path.Join(p.Root, TsconfigJSON), Mode: protocol.StandardFile,
		Content: ProjectTsconfig(p.Root), Reason: reasonTsconfig,
	}}
	if build, err := os.ReadFile(filepath.Join(dir, TsconfigBuild)); err == nil && ReplaceableBuildTsconfig(build) {
		out = append(out, protocol.StandardFileContent{
			Project: p.Name, Path: path.Join(p.Root, TsconfigBuild), Mode: protocol.StandardFile,
			Content: BuildTsconfig(testSuffix), Reason: reasonTsconfig,
		})
	}
	return out
}
