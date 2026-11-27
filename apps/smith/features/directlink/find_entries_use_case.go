package directlink

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// entryNames are the names smith has in a folder of entries, by system.
func entryNames(goos string) []string {
	if goos == "windows" {
		return []string{"smith.cmd", "smith.ps1", "smith", "smith.exe"}
	}
	return []string{"smith"}
}

// ownedDirs are the folders where the package manager that installed exe puts
// its entries: the npm prefix (and the bin folder beside a Unix lib folder), and
// the Scripts or bin folder of the Python that holds site-packages. Only these
// are looked at: another smith on PATH, with its own install, is never touched,
// and smith never takes an entry for its own because the name matches.
func ownedDirs(exe string) []string {
	p := filepath.ToSlash(exe)
	lower := strings.ToLower(p)
	var dirs []string
	if i := strings.Index(lower, "/node_modules/monosmith/"); i >= 0 {
		prefix := filepath.FromSlash(p[:i])
		dirs = append(dirs, prefix, filepath.Join(filepath.Dir(prefix), "bin"))
	}
	// pnpm keeps the program in its store and the entries in its home: <home>/bin
	// (pnpm 11 and later) or <home> itself.
	if i := strings.Index(lower, "/store/v"); i >= 0 && strings.Contains(lower, "/links/@monosmith/") {
		home := filepath.FromSlash(p[:i])
		dirs = append(dirs, filepath.Join(home, "bin"), home)
	}
	if i := strings.Index(lower, "/site-packages/"); i >= 0 {
		root := filepath.FromSlash(p[:i])
		for range 4 {
			dirs = append(dirs, filepath.Join(root, "Scripts"), filepath.Join(root, "bin"))
			root = filepath.Dir(root)
		}
	}
	return dirs
}

// FindEntries lists the smith entries of the installation whose program is exe,
// each with its kind.
//
// Parameters: exe is the installed program; goos is runtime.GOOS; runDir is the
// folder of the links an entry already changed calls; target is the program an
// entry of this version should call; pathDirs are more folders to look in (the
// PATH folders, for a program the launchers installed), where only an entry that
// names the smith package or calls the run folder counts.
// Returns: the entries, without duplicates.
func FindEntries(exe, goos, runDir, target string, pathDirs []string) []Entry {
	var found []Entry
	var seen []string
	for _, dir := range append(ownedDirs(exe), pathDirs...) {
		if dir == "" {
			continue
		}
		for _, name := range entryNames(goos) {
			file := filepath.Join(dir, name)
			key := strings.ToLower(file)
			if _, err := os.Lstat(file); err != nil || slices.Contains(seen, key) {
				continue
			}
			seen = append(seen, key)
			if kind, ok := classify(file, exe, runDir, target); ok {
				found = append(found, Entry{Path: file, Kind: kind})
			}
		}
	}
	return found
}

// isNativeEntry is true when the installed program is the entry itself: the
// install script puts it in ~/.local/bin. No launcher stands in front of it, so
// there is nothing to change. A program in the versions folder of home is not: a
// launcher package starts it.
func isNativeEntry(home, exe string) bool {
	name := strings.ToLower(filepath.Base(exe))
	return (name == "smith" || name == "smith.exe") && len(ownedDirs(exe)) == 0 && !Managed(home, exe)
}
