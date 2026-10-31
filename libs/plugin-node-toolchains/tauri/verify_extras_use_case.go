package tauri

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VerifyExtras checks the built app against the declared sidecars and
// resources: Tauri puts each next to the app's program in the target's release
// folder, and into the bundles that can be read on disk: a .deb's data folder,
// an AppImage's AppDir, a macOS .app, and on Windows the files an .msi
// installs (read with msiexec /a, which extracts without installing). It fails
// when one is missing from a place it looked. What it cannot open (the NSIS
// .exe, .dmg, .rpm, the .AppImage file, and an .msi away from Windows) is named
// in the lines it returns, so a log says what was not checked.
func VerifyExtras(workspaceRoot, projectRoot string, extras Extras) ([]string, error) {
	release := filepath.Join(workspaceRoot, filepath.FromSlash(TargetDir(projectRoot)), "release")
	triple, _ := HostTriple()
	suffix := exeSuffix(triple)
	var lines, missing []string
	places := append([]string{release}, unpackedBundles(release)...)
	for _, s := range extras.Sidecars {
		for _, dir := range places {
			if !fileExists(filepath.Join(dir, s.Name+suffix)) {
				missing = append(missing, fmt.Sprintf("%s is not in %s", s.Name, rel(workspaceRoot, dir)))
			}
		}
		lines = append(lines, fmt.Sprintf("sidecar %s: checked in %d place(s)", s.Name, len(places)))
	}
	// Tauri copies the resources next to the app's program as well.
	staged := filepath.Join(workspaceRoot, filepath.FromSlash(PrepareDir(projectRoot)), "resources")
	var resources []string
	_ = filepath.WalkDir(staged, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			r, _ := filepath.Rel(staged, p)
			resources = append(resources, r)
		}
		return nil
	})
	for _, r := range resources {
		if !fileExists(filepath.Join(release, r)) {
			missing = append(missing, fmt.Sprintf("the resource %s is not in %s", filepath.ToSlash(r), rel(workspaceRoot, release)))
		}
	}
	if len(resources) > 0 {
		lines = append(lines, fmt.Sprintf("resources: %d file(s) checked", len(resources)))
	}
	msis, _ := filepath.Glob(filepath.Join(release, "bundle", "msi", "*.msi"))
	for _, msi := range msis {
		if triple == "" || exeSuffix(triple) == "" {
			break
		}
		gone, err := msiMissing(msi, append(sidecarFiles(extras, suffix), resources...))
		if err != nil {
			lines = append(lines, "msi not opened: "+err.Error())
			continue
		}
		for _, g := range gone {
			missing = append(missing, fmt.Sprintf("%s is not in %s", g, rel(workspaceRoot, msi)))
		}
		lines = append(lines, "msi: "+filepath.Base(msi)+" opened and checked")
	}
	lines = append(lines, "not opened: the NSIS .exe, .dmg, .rpm, the .AppImage file (Tauri itself stops the build when a declared file is missing)")
	if len(missing) > 0 {
		return lines, fmt.Errorf("the built app lacks what smith.toml declares: %s", strings.Join(missing, "; "))
	}
	return lines, nil
}

func sidecarFiles(extras Extras, suffix string) []string {
	var names []string
	for _, s := range extras.Sidecars {
		names = append(names, s.Name+suffix)
	}
	return names
}

// unpackedBundles are the folders of a build's bundles where a sidecar sits
// beside the app's program; a build leaves only some of them.
func unpackedBundles(release string) []string {
	var dirs []string
	for _, pattern := range []string{"bundle/macos/*.app/Contents/MacOS", "bundle/deb/*/data/usr/bin", "bundle/appimage/*.AppDir/usr/bin"} {
		found, _ := filepath.Glob(filepath.Join(release, filepath.FromSlash(pattern)))
		dirs = append(dirs, found...)
	}
	return dirs
}

// msiMissing extracts an .msi into a temporary folder (msiexec /a installs
// nothing) and returns the wanted files, as slash paths, that it does not hold.
func msiMissing(msi string, wanted []string) ([]string, error) {
	if _, err := exec.LookPath("msiexec"); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "smith-msi-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if out, err := exec.Command("msiexec", "/a", msi, "/qn", "TARGETDIR="+dir).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("msiexec: %w %s", err, out)
	}
	have := map[string]bool{}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			have[filepath.Base(p)] = true
		}
		return nil
	})
	var gone []string
	for _, w := range wanted {
		if !have[filepath.Base(w)] {
			gone = append(gone, filepath.ToSlash(w))
		}
	}
	return gone, nil
}
