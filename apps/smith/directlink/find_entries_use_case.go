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
// entry of this version should call.
// Returns: the entries, without duplicates.
func FindEntries(exe, goos, runDir, target string) []Entry {
	var found []Entry
	var seen []string
	for _, dir := range ownedDirs(exe) {
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
