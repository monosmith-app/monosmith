package releasing

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/releaseassets"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// BuildAssets publishes an app for each platform of its release as a
// self-contained single file (the .NET runtime is inside, so the program runs
// where .NET is not installed), and zips the result with the README and
// licence files, one zip per platform, with the checksums of them all. A .NET
// program is managed code, so one machine publishes every platform: there is
// no linker and no system library to cross-build against, only each
// platform's runtime pack, which NuGet supplies. The version is stamped with
// -p:Version.
func BuildAssets(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	spec := *params.Assets
	p := params.Project
	if len(spec.Platforms) == 0 {
		return protocol.PublishResult{}, fmt.Errorf("%s: no platforms to build assets for", p.Name)
	}
	file, err := projectFile(workspaceRoot, p)
	if err != nil {
		return protocol.PublishResult{}, err
	}
	assembly := assemblyName(workspaceRoot, p, file)
	type target struct{ platform, rid, archive string }
	var targets []target
	for _, platform := range spec.Platforms {
		goos, goarch, ok := strings.Cut(platform, "/")
		if !ok || goos == "" || goarch == "" {
			return protocol.PublishResult{}, fmt.Errorf("%s: platform %q is not os/arch", p.Name, platform)
		}
		rid, err := RuntimeIdentifier(platform)
		if err != nil {
			return protocol.PublishResult{}, fmt.Errorf("%s: %w", p.Name, err)
		}
		name, err := releaseassets.AssetName(spec, p.Name, params.Version, goos, goarch, "zip")
		if err != nil {
			return protocol.PublishResult{}, err
		}
		targets = append(targets, target{platform, rid, name})
	}

	result := protocol.PublishResult{Published: !params.DryRun, AssetPlatforms: map[string]string{}, Binary: assembly}
	outDir := filepath.Join(workspaceRoot, filepath.FromSlash(spec.OutDir))
	if !params.DryRun {
		if err := os.RemoveAll(outDir); err != nil {
			return protocol.PublishResult{}, err
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	var archives []string
	for _, t := range targets {
		asset := path.Join(spec.OutDir, t.archive)
		result.Assets = append(result.Assets, asset)
		result.AssetPlatforms[asset] = t.platform
		if params.DryRun {
			continue
		}
		staging := filepath.Join(outDir, t.rid)
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		toolrun.Step("Publishing " + p.Name + " for " + t.platform)
		argv := []string{"dotnet", "publish", file, "--nologo", "-c", "Release", "-r", t.rid, "--self-contained", "true",
			"-p:PublishSingleFile=true", "-p:IncludeNativeLibrariesForSelfExtract=true", "-p:DebugType=None", "-p:Version=" + params.Version, "-o", staging}
		if err := DotnetRunner(dir, argv); err != nil {
			return protocol.PublishResult{}, fmt.Errorf("publishing %s: %w", t.platform, err)
		}
		entries, err := publishedEntries(staging, assembly)
		if err != nil {
			return protocol.PublishResult{}, fmt.Errorf("%s: %w", t.platform, err)
		}
		entries = addCompanions(entries, releaseassets.Companions(workspaceRoot, p.Root))
		archive := filepath.Join(outDir, t.archive)
		if err := releaseassets.WriteZip(archive, entries); err != nil {
			return protocol.PublishResult{}, fmt.Errorf("%s: %w", t.platform, err)
		}
		archives = append(archives, archive)
		if err := os.RemoveAll(staging); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	result.Assets = append(result.Assets, path.Join(spec.OutDir, releaseassets.ChecksumsFile))
	if !params.DryRun {
		if err := releaseassets.WriteChecksums(filepath.Join(outDir, releaseassets.ChecksumsFile), archives...); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	return result, nil
}

// assemblyName is the name of the program the project builds.
func assemblyName(workspaceRoot string, p protocol.Project, file string) string {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), file))
	if err != nil {
		return strings.TrimSuffix(file, filepath.Ext(file))
	}
	parsed, err := discovery.ParseProject(path.Join(p.Root, file), src)
	if err != nil {
		return strings.TrimSuffix(file, filepath.Ext(file))
	}
	return parsed.AssemblyName
}

// publishedEntries lists what dotnet publish wrote: every file but the
// debugging symbols, the program marked executable.
func publishedEntries(dir, assembly string) ([]releaseassets.Entry, error) {
	var entries []releaseassets.Entry
	program := false
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(d.Name(), ".pdb") {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		isProgram := d.Name() == assembly || d.Name() == assembly+".exe"
		program = program || isProgram
		entries = append(entries, releaseassets.Entry{Path: p, Name: filepath.ToSlash(rel), Executable: isProgram})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !program {
		return nil, fmt.Errorf("dotnet publish wrote no program named %s: is the project an app (OutputType Exe)?", assembly)
	}
	return entries, nil
}

// addCompanions adds the README and licence files an app's own output does
// not already have.
func addCompanions(entries, companions []releaseassets.Entry) []releaseassets.Entry {
	for _, c := range companions {
		have := false
		for _, e := range entries {
			if strings.EqualFold(e.Name, c.Name) {
				have = true
			}
		}
		if !have {
			entries = append(entries, c)
		}
	}
	return entries
}
