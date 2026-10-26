package legacyjesthelpers

import (
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Cleanup is what the older helper files leave to do in a workspace.
type Cleanup struct {
	// Files are the helper files convert wrote that are still at the root.
	Files []string
	// Fix removes them and points the configs at the package; nil when
	// something named a helper in a way the cleanup does not recognise.
	Fix *protocol.Fix
	// Manual says what needs a hand when Fix is nil.
	Manual string
}

// FindCleanup looks for the helper files at the workspace root and plans
// their removal. projectDirs are the workspace-relative directories of the
// projects whose Jest configs may name the asset transformer. The plan is
// all or nothing: deleting a helper while something still loads it would
// break every package's tests, so one thing unrecognised leaves Fix nil.
func FindCleanup(root string, projectDirs []string) Cleanup {
	var found Cleanup
	read := func(rel string) (string, bool) {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		return string(raw), err == nil
	}
	var deletes []string
	if text, ok := read(ResolverFile); ok && OwnsResolver(text) {
		found.Files = append(found.Files, ResolverFile)
		deletes = append(deletes, ResolverFile)
	}
	if text, ok := read(AssetTransformFile); ok && OwnsAssetTransform(text) {
		found.Files = append(found.Files, AssetTransformFile)
		deletes = append(deletes, AssetTransformFile)
	}
	if len(found.Files) == 0 {
		return found
	}

	edits := map[string]string{}
	var unrecognised []string
	for _, rel := range jestFiles(root, append([]string{"."}, projectDirs...)) {
		text, _ := read(rel)
		if !NamesHelper(text) {
			continue
		}
		out := text
		if strings.HasPrefix(path.Base(rel), "jest.preset.") {
			out, _ = RewritePreset(out)
		}
		out, _ = RewriteConfig(out)
		if NamesHelper(out) {
			unrecognised = append(unrecognised, rel)
			continue
		}
		edits[rel] = out
	}
	if len(unrecognised) > 0 {
		found.Manual = "still names " + strings.Join(found.Files, " or ") + " in a way smith cannot rewrite: " + strings.Join(unrecognised, ", ") +
			". Make them use " + Package + " (preset: '" + Package + "', transform '" + Package + "/asset-transform'), then delete the helper files"
		return found
	}

	manifest, ok := read("package.json")
	if !ok {
		found.Manual = "the workspace has no root package.json to add " + Package + " to: npm i -D " + Package + ", then delete " + strings.Join(found.Files, ", ")
		return found
	}
	withPackage, err := addDevDependency(manifest)
	if err != nil {
		found.Manual = "package.json could not be edited (" + err.Error() + "): npm i -D " + Package + ", then delete " + strings.Join(found.Files, ", ")
		return found
	}

	fix := &protocol.Fix{
		Description: "replace the Jest helper files convert wrote (" + strings.Join(found.Files, ", ") + ") with " + Package,
		Steps:       []protocol.FixStep{{Description: "install " + Package, Argv: []string{"npm", "install", "--no-audit", "--no-fund"}}},
	}
	for _, rel := range slices.Sorted(maps.Keys(edits)) {
		fix.Files = append(fix.Files, protocol.FileEdit{Path: rel, Content: edits[rel]})
	}
	if withPackage != manifest {
		fix.Files = append(fix.Files, protocol.FileEdit{Path: "package.json", Content: withPackage})
	}
	for _, rel := range deletes {
		fix.Files = append(fix.Files, protocol.FileEdit{Path: rel, Delete: true})
	}
	found.Fix = fix
	return found
}

// jestFiles are the Jest configs and presets at the top of each directory.
func jestFiles(root string, dirs []string) []string {
	var out []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && (strings.HasPrefix(e.Name(), "jest.config.") || strings.HasPrefix(e.Name(), "jest.preset.")) {
				out = append(out, path.Join(dir, e.Name()))
			}
		}
	}
	return out
}

// addDevDependency lists the package in the root devDependencies, unless some
// dependency section already does.
func addDevDependency(manifest string) (string, error) {
	doc, err := jsondoc.ParseDocument([]byte(manifest))
	if err != nil {
		return "", err
	}
	for _, section := range []string{"devDependencies", "dependencies"} {
		if deps := doc.Object(section, false); deps != nil && deps.Has(Package) {
			return manifest, nil
		}
	}
	doc.Object("devDependencies", true).SetString(Package, PackageRange)
	return string(doc.Bytes()), nil
}
