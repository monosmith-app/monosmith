package adoptsplit

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// moveBuildIntoDist points the build of an app (or a VS Code extension) at
// the workspace's dist/<project dir>, the folder smith builds and caches it
// into, so the workspace does not start with a failing doctor check. A library
// or package keeps building inside its own folder (ADR 0102). It uses the
// same edit as smith doctor --fix (the tsconfig outDir becomes an --outDir
// flag of the build script, and main, types and start follow). It leaves the
// build alone, and says why, when that edit is refused. It returns the note
// for the person, or "" when there is nothing to move.
func moveBuildIntoDist(req Request) (string, error) {
	dir := filepath.Join(req.Root, filepath.FromSlash(req.Dir))
	if raw, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		var pkg discovery.PackageManifest
		if json.Unmarshal(raw, &pkg) == nil {
			fallback := protocol.ProjectPackage
			if pkg.Private {
				fallback = protocol.ProjectLib
			}
			if !discovery.BuildsLikeApp(projectkind.Resolve(pkg.Smith.Kind, req.Dir, fallback), pkg.Engines.VSCode != "") {
				return "", nil
			}
		}
	}
	config, outDir, ok := manifestfixes.LocateOutDir(dir)
	if !ok {
		return "", nil
	}
	p := protocol.Project{Name: path.Base(req.Dir), Root: req.Dir}
	if protocol.InBuildDir(path.Clean(path.Join(req.Dir, filepath.ToSlash(outDir)))) {
		return "", nil
	}
	if reason := manifestfixes.MigrationRefusal(req.Root, p, config, outDir, nil); reason != "" {
		return "the build stays where it is, because " + reason + ".", nil
	}
	fix := manifestfixes.BuildIntoDistFix(req.Root, p, config, outDir)
	if fix == nil {
		return "", nil
	}
	for _, f := range fix.Files {
		if err := os.WriteFile(filepath.Join(req.Root, filepath.FromSlash(f.Path)), []byte(f.Content), 0o644); err != nil {
			return "", err
		}
	}
	return "the build writes into " + protocol.ProjectBuildDir(req.Dir) + ", where smith caches it; main, types and the start script follow.", nil
}
